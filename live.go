package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	"golang.org/x/sync/errgroup"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func initLive(app core.App) {
	rooms := NewLiveRooms(app)
	tname := "live_rooms"
	app.OnBeforeServe().Add(func(e *core.ServeEvent) (err error) {
		defer err0.Then(&err, func() {
			app.Logger().Info("启动时批量推流成功")
		}, func() {
			app.Logger().Error("启动时批量推流出错了", "error", err)
			err = nil
		})
		dao := app.Dao()
		records := try.To1(dao.FindRecordsByExpr(tname))
		eg := new(errgroup.Group)
		for _, _r := range records {
			r := _r
			eg.Go(func() error {
				if push := r.GetBool("push"); !push {
					r.Set("running", false)
					return dao.Save(r)
				}
				return rooms.Start(r)
			})
		}
		try.To(eg.Wait())
		return nil
	})
	app.OnRecordAfterCreateRequest(tname).Add(func(e *core.RecordCreateEvent) error {
		ctx := e.HttpContext.Request().Context()
		return rooms.HandleChange(ctx, e.Record, e.UploadedFiles)
	})
	app.OnRecordAfterUpdateRequest(tname).Add(func(e *core.RecordUpdateEvent) error {
		ctx := e.HttpContext.Request().Context()
		return rooms.HandleChange(ctx, e.Record, e.UploadedFiles)
	})
	app.OnRecordAfterDeleteRequest(tname).Add(func(e *core.RecordDeleteEvent) error {
		return rooms.Stop(e.Record)
	})
}

type LiveRooms struct {
	core.App
	pool   map[string]*LiveRoom
	locker sync.Locker
}

func NewLiveRooms(app core.App) *LiveRooms {
	return &LiveRooms{
		App:    app,
		pool:   map[string]*LiveRoom{},
		locker: &sync.Mutex{},
	}
}

func (rooms *LiveRooms) HandleChange(ctx context.Context, record *models.Record, filesMap map[string][]*filesystem.File) (err error) {
	defer err0.Then(&err, nil, nil)
	push := record.GetBool("push")
	if !push {
		return rooms.Stop(record)
	}
	room := rooms.GetRoom(record.Id)
	if room != nil {
		try.To(rooms.Stop(record))
	}
	if _, ok := filesMap["video"]; ok { // 有新视频时将偏移量设为0
		record.Set("offset", 0)
	}
	return rooms.Start(record)
}

func (rooms *LiveRooms) Start(record *models.Record) (err error) {
	logger := rooms.Logger().With(
		"id", record.Id,
		"name", record.GetString("name"),
	)
	defer err0.Then(&err, func() {
		logger.Info("启动推流成功")
	}, func() {
		logger.Error("启动推流失败", "error", err)
	})

	rooms.locker.Lock()
	defer rooms.locker.Unlock()

	video := record.GetString("video")
	basedir := record.BaseFilesPath()
	basedir = filepath.Join(rooms.App.DataDir(), "storage", basedir)
	video = filepath.Join(basedir, video)

	ctx := context.Background()
	dao := rooms.Dao()

	duration := int(try.To1(ffprobe.ProbeURL(ctx, video)).Format.DurationSeconds)
	record.Set("running", false)
	record.Set("duration", duration)
	try.To(dao.Save(record))

	server := record.GetString("rtmps")
	offset := record.GetInt("offset")
	room := NewLiveRoom(ctx, video, server, offset)
	rooms.pool[record.Id] = room
	{
		cmd := strings.Join(room.Cmd.Args, " ")
		logger.Debug("推流指令", "cmd", cmd)
	}

	try.To(room.Start())
	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
			t := time.NewTicker(time.Second)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					offset = (offset + 1) % duration
					record.Set("offset", offset)
					if err := dao.Save(record); err != nil {
						logger.Warn("更新时间偏移量失败")
					}
				}
			}
		}()
		room.Wait()
	}()

	record.Set("running", true)
	try.To(dao.Save(record))

	return
}

func (rooms *LiveRooms) GetRoom(id string) *LiveRoom {
	rooms.locker.Lock()
	defer rooms.locker.Unlock()
	room, ok := rooms.pool[id]
	if !ok || room == nil {
		return nil
	}
	return room
}

func (rooms *LiveRooms) Stop(record *models.Record) (err error) {
	logger := rooms.Logger().With(
		"id", record.Id,
		"name", record.GetString("name"),
	)
	defer err0.Then(&err, func() {
		logger.Info("停止推流成功")
	}, func() {
		logger.Error("停止推流时出错", "error", err)
	})

	id := record.Id
	room := rooms.GetRoom(record.Id)
	if room == nil {
		logger.Warn("正在停止未在运行的推流")
		return nil
	}

	rooms.locker.Lock()
	defer rooms.locker.Unlock()

	if err := room.Close(); err != nil {
		logger.Warn("关闭推流进程出错", "error", err)
	}

	record.Set("offset", 0) // 主动退出时, 将 offset 设为 0
	record.Set("running", false)
	try.To(rooms.Dao().Save(record))
	rooms.pool[id] = nil

	return
}

type LiveRoom struct {
	*exec.Cmd
	running bool
}

func NewLiveRoom(ctx context.Context, input, server string, offset int) *LiveRoom {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-re",                // 推流
		"-stream_loop", "-1", // 无限循环
		"-ss", fmt.Sprintf("%d", offset), // 时间偏移, 当意外重启时可以快速恢复
		"-i", input,
		// "-vcodec", "libx264", "-acodec", "aac", //转码, 不是必须的
		"-f", "flv", server,
	)
	return &LiveRoom{
		Cmd: cmd,
	}
}

var _ io.Closer = (*LiveRoom)(nil)

func (room *LiveRoom) Close() error {
	return room.Process.Kill()
}
