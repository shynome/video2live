package main

import (
	"context"
	"os/exec"
	"strconv"
	"strings"

	"github.com/shynome/err0/try"
)

func tryGetVideoDuration(ctx context.Context, input string) int {
	cmd := exec.CommandContext(ctx, "ffprobe", "-i", input, "-show_format")
	output := try.To1(cmd.Output())
	info := strings.Split(string(output), "\n")
	for _, s := range info {
		if !strings.HasPrefix(s, "duration=") {
			continue
		}
		parts := strings.Split(s, "=")
		d := try.To1(strconv.Atoi(parts[1]))
		return d
	}
	return 0
}
