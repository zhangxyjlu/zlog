# zlog 使用说明

## zlog 是什么？

zlog 是封装 uber-go/zap 的 go 语言日志组件。Zap 是一个高性能的结构化日志记录库，特别适用于 Go 语言的应用程序。
它是由 Uber 开发并维护的，旨在提供极佳的性能和低分配成本。

相比于 [zap](https://github.com/uber-go/zap)，提供了日志切分、配置文件读取、配置文件动态加载等功能。

其中日志切分使用的是 [lumberjack](https://github.com/natefinch/lumberjack) 组件实现的。

## 配置说明

在项目根目录下建立 `resources/config.yaml` 文件。当前支持如下字段配置
```yaml
# 日志配置
Level: info         # 日志级别，默认info
# 日志输出编码方式，支持 console、json。
# console 即通常见到的标准格式：2023-10-10T18:30:26.438+0800    INFO    cmd/main.go:15  This is a Info message.
# json 对日志进行json输出：{"level":"ERROR","ts":"2023-10-10T19:36:56.025+0800","caller":"cmd/main.go:18","msg":"This is a Error message."}
EncoderType: console 
Path: /var/log/     # 日志文件存放目录，默认/var/log/{运行文件名}；如运行文件为hello/hello.exe，则path为：/var/log/hello
FileName: root.log  # 日志文件名称；默认是 root.log
MaxSize: 20         # 单文件最大存储，单位MB，默认20M
MaxBackups: 2       # 保留个数 默认100个
MaxAge: 7           # 最多保留的旧日志文件的天数，默认7 days
LocalTime: false    # 是否使用本地时间格式，默认false
Compress: true      # 是否启用日志文件的压缩，默认压缩
OutMod: console     # 输出模式，支持：console——输出到控制台；both——文件+控制台都输出；file——输出到文件；默认输出到控制台
```
## 使用方法

### 基本用法

zlog 中封装了 zap 的基本使用方法，如 Info、Infof、Infow等。如想使用更多的 zap 日志工能。可以调用 `zlog.Logger`。
```go
import (
	"github.com/zhangxyjlu/zlog"
	"time"
)

func main() {
	println("hello log")
	defer zlog.Sync()
	i := 1
	for {
		zlog.Errorf("第 %v 次", i)
		zlog.Debug("This is a Debug message.")
		zlog.Info("This is a Info message.")
		zlog.Warn("This is a Warn message.")
		zlog.Error("This is a Error message.")
		i++
		time.Sleep(500 * time.Millisecond) // 每隔0.5秒钟打印一次日志
	}
}
```

### 支持自定义配置获取 Logger 对象

默认的Logger对象和配置文件绑定，且监听配置文件修改。可以通过 `zlog.GetLogger(Conf)` 获取不同的日志配置，根据逻辑输出到不同的文件等。
```go
	Conf := &zlog.LogConfig{
		Level:       "info",
		EncoderType: "console",
		Path:        "var/log",
		FileName:    "access.log",
	}
	Logger := zlog.GetLogger(Conf)
	defer Logger.Sync()
	Logger.Info("test")
	
```

## 动态配置

支持热加载配置文件，程序运行时，可以动态修改配置文件，热生效。如修改日志级别等。

## 其他用法

参考 [zap 文档](https://pkg.go.dev/go.uber.org/zap)




