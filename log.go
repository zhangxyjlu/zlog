package zlog

import (
	"github.com/fsnotify/fsnotify"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const DefaultLogPath = "/var/log"

var Logger *zap.SugaredLogger

var logLevel = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
}

var watchOnce = sync.Once{}

type LogConfig struct {
	Level       string
	EncoderType string
	Path        string
	FileName    string
	MaxSize     int
	MaxBackups  int
	MaxAge      int
	LocalTime   bool
	Compress    bool
	OutMod      string
}

func init() {
	var conf *LogConfig
	var err error
	if conf, err = loadConfig(); err != nil {
		conf = getDefaultConf()
	}
	Logger = GetLogger(conf)
}

func getDefaultConf() *LogConfig {
	var defaultConf = &LogConfig{
		Level:       "info",
		EncoderType: "console",
		Path:        DefaultLogPath,
		FileName:    "root.log",
		MaxSize:     20,
		MaxBackups:  0,
		MaxAge:      7,
		LocalTime:   false,
		Compress:    true,
		OutMod:      "console",
	}
	exePath, err := os.Executable()
	if err != nil {
		return defaultConf
	}
	// 获取运行文件名称，作为/var/log目录下的子目录
	serviceName := strings.TrimSuffix(filepath.Base(exePath), filepath.Ext(filepath.Base(exePath)))
	defaultConf.Path = filepath.Join(DefaultLogPath, serviceName)
	return defaultConf
}

func GetLogger(conf *LogConfig) *zap.SugaredLogger {
	writeSyncer := getLogWriter(conf)
	encoder := getEncoder(conf)
	level, ok := logLevel[strings.ToLower(conf.Level)]
	if !ok {
		level = logLevel["info"]
	}
	core := zapcore.NewCore(encoder, writeSyncer, level)
	logger := zap.New(core, zap.AddCaller())
	return logger.Sugar()
}

func loadConfig() (*LogConfig, error) {
	viper.AddConfigPath("./resources")
	// 设置要读取的配置文件名（例如：config.yaml）
	viper.SetConfigName("config")
	// 设置配置文件类型（例如：yaml、json）
	viper.SetConfigType("yaml")
	config, err := parseConfig()
	if err != nil {
		return nil, err
	}
	watchConfig()
	return config, nil
}

func watchConfig() {
	// 监听配置文件的变化
	watchOnce.Do(func() {
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			Logger.Warn("Config file changed")
			// 重新加载配置
			conf, err := parseConfig()
			if err != nil {
				Logger.Warnf("Error reloading config file: %v\n", err)
			} else {
				Logger = GetLogger(conf)
			}
		})
	})
}

func parseConfig() (*LogConfig, error) {
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	var config LogConfig
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// //获取编码器,NewJSONEncoder()输出json格式，NewConsoleEncoder()输出普通文本格式
func getEncoder(conf *LogConfig) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	//指定时间格式 for example: 2021-09-11t20:05:54.852+0800
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	//按级别显示不同颜色，不需要的话取值zapcore.CapitalLevelEncoder就可以了
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// 显示完整文件路径
	//EncoderConfig.EncodeCaller=zapcore.FullCallerEncoder
	//NewJSONEncoder()输出json格式，NewConsoleEncoder()输出普通文本格式
	if strings.ToLower(conf.EncoderType) == "json" {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(conf *LogConfig) zapcore.WriteSyncer {
	// 只输出到控制台
	if conf.OutMod == "console" {
		return zapcore.AddSync(os.Stdout)
	}
	// 日志文件配置
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filepath.Join(conf.Path, conf.FileName),
		MaxSize:    conf.MaxSize,
		MaxBackups: conf.MaxBackups,
		MaxAge:     conf.MaxAge,
		LocalTime:  conf.LocalTime,
		Compress:   conf.Compress,
	}
	if conf.OutMod == "both" {
		// 控制台和文件都输出
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumberJackLogger), zapcore.AddSync(os.Stdout))
	}
	if conf.OutMod == "file" {
		// 只输出到文件
		return zapcore.AddSync(lumberJackLogger)
	}
	return zapcore.AddSync(os.Stdout)
}

//以下封装 SugaredLogger 常用接口

// With adds a variadic number of fields to the logging context. It accepts a
// mix of strongly-typed Field objects and loosely-typed key-value pairs. When
// processing pairs, the first element of the pair is used as the field key
// and the second as the field value.
func With(args ...interface{}) *zap.SugaredLogger {
	return Logger.With(args...)
}

// Debug logs the provided arguments at [DebugLevel].
// Spaces are added between arguments when neither is a string.
func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

// Info logs the provided arguments at [].
// Spaces are added between arguments when neither is a string.
func Info(args ...interface{}) {
	Logger.Info(args...)
}

// Warn logs the provided arguments at [WarnLevel].
// Spaces are added between arguments when neither is a string.
func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

// Error logs the provided arguments at [ErrorLevel].
// Spaces are added between arguments when neither is a string.
func Error(args ...interface{}) {
	Logger.Error(args...)
}

// DPanic logs the provided arguments at [DPanicLevel].
// In development, the logger then panics. (See [DPanicLevel] for details.)
// Spaces are added between arguments when neither is a string.
func DPanic(args ...interface{}) {
	Logger.DPanic(args...)
}

// Panic constructs a message with the provided arguments and panics.
// Spaces are added between arguments when neither is a string.
func Panic(args ...interface{}) {
	Logger.Panic(args...)
}

// Fatal constructs a message with the provided arguments and calls os.Exit.
// Spaces are added between arguments when neither is a string.
func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

// Debugf formats the message according to the format specifier
// and logs it at [DebugLevel].
func Debugf(template string, args ...interface{}) {
	Logger.Debugf(template, args...)
}

// Infof formats the message according to the format specifier
// and logs it at [].
func Infof(template string, args ...interface{}) {
	Logger.Infof(template, args...)
}

// Warnf formats the message according to the format specifier
// and logs it at [WarnLevel].
func Warnf(template string, args ...interface{}) {
	Logger.Warnf(template, args...)
}

// Errorf formats the message according to the format specifier
// and logs it at [ErrorLevel].
func Errorf(template string, args ...interface{}) {
	Logger.Errorf(template, args...)
}

// DPanicf formats the message according to the format specifier
// and logs it at [DPanicLevel].
// In development, the logger then panics. (See [DPanicLevel] for details.)
func DPanicf(template string, args ...interface{}) {
	Logger.DPanicf(template, args...)
}

// Panicf formats the message according to the format specifier
// and panics.
func Panicf(template string, args ...interface{}) {
	Logger.Panicf(template, args...)
}

// Fatalf formats the message according to the format specifier
// and calls os.Exit.
func Fatalf(template string, args ...interface{}) {
	Logger.Fatalf(template, args...)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//
//	s.With(keysAndValues).Debug(msg)
func Debugw(msg string, keysAndValues ...interface{}) {
	Logger.Debugw(msg, keysAndValues...)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Infow(msg string, keysAndValues ...interface{}) {
	Logger.Infow(msg, keysAndValues...)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Warnw(msg string, keysAndValues ...interface{}) {
	Logger.Warnw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Errorw(msg string, keysAndValues ...interface{}) {
	Logger.Errorw(msg, keysAndValues...)
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func DPanicw(msg string, keysAndValues ...interface{}) {
	Logger.DPanicw(msg, keysAndValues...)
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func Panicw(msg string, keysAndValues ...interface{}) {
	Logger.Panicw(msg, keysAndValues...)
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func Fatalw(msg string, keysAndValues ...interface{}) {
	Logger.Fatalw(msg, keysAndValues...)
}

// Debugln logs a message at [DebugLevel].
// Spaces are always added between arguments.
func Debugln(args ...interface{}) {
	Logger.Debugln(args...)
}

// Infoln logs a message at [].
// Spaces are always added between arguments.
func Infoln(args ...interface{}) {
	Logger.Infoln(args...)
}

// Warnln logs a message at [WarnLevel].
// Spaces are always added between arguments.
func Warnln(args ...interface{}) {
	Logger.Warnln(args...)
}

// Errorln logs a message at [ErrorLevel].
// Spaces are always added between arguments.
func Errorln(args ...interface{}) {
	Logger.Errorln(args...)
}

// DPanicln logs a message at [DPanicLevel].
// In development, the logger then panics. (See [DPanicLevel] for details.)
// Spaces are always added between arguments.
func DPanicln(args ...interface{}) {
	Logger.DPanicln(args...)
}

// Panicln logs a message at [PanicLevel] and panics.
// Spaces are always added between arguments.
func Panicln(args ...interface{}) {
	Logger.Panicln(args...)
}

// Fatalln logs a message at [FatalLevel] and calls os.Exit.
// Spaces are always added between arguments.
func Fatalln(args ...interface{}) {
	Logger.Fatalln(args...)
}

// Sync flushes any buffered log entries.
func Sync() error {
	return Logger.Sync()
}
