## 简介

将一个视频无限循环推流到直播间中

## 使用

1. 启动服务

```sh
docker run --restart always --name video2live -v $PWD/video2live/:/app/pb_data/ -p 8090:8090 shynome/video2live:v0.0.1
```

2. 访问 `http://127.0.0.1:8090/`, 设置好帐号密码, 新建 `live_rooms` 即可开始推流
   ![](images/图解.png)

## 特色

- 重启后可以接着上次播放进度继续推流(主动停止则会从头推流)
- 运行正常时 `offset` 会每秒 +1
