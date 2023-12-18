package main

import (
	"context"
	"testing"

	"github.com/shynome/err0/try"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func TestGetVideoDuration(t *testing.T) {
	ctx := context.Background()
	d := try.To1(ffprobe.ProbeURL(ctx, "test.mp4"))
	t.Log(d)
}
