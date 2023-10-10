package main

import (
	"github.com/zhangxyjlu/zlog"
	"time"
)

func main() {
	println("hello log")
	defer zlog.Sync()
	i := 1
	j := 1
	for {
		zlog.Errorf("第 %d 次, 开始计数 %d", i, j)
		zlog.Debug("This is a Debug message.")
		zlog.Info("This is a Info message.")
		zlog.Warn("This is a Warn message.")
		zlog.Error("This is a Error message.")
		i++
		j++
		time.Sleep(500 * time.Millisecond) // 每隔0.5秒钟打印一次日志
	}
}
