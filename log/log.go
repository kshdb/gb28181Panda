package log

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"time"
)

// Logger interface used as base logger throughout the library.
type Logger interface {
	Print(args ...interface{})
	Printf(format string, args ...interface{})

	Trace(args ...interface{})
	Tracef(format string, args ...interface{})

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	Panic(args ...interface{})
	Panicf(format string, args ...interface{})

	WithPrefix(prefix string) Logger
	Prefix() string

	WithFields(fields Fields) Logger
	Fields() Fields

	SetLevel(level Level)
}

type Loggable interface {
	Log() Logger
}

var (
	l = initLog()
)

func initLog() *zap.SugaredLogger {
	core := zapcore.NewCore(getJSONEncoder(), getLoggerWrite(), getLoggerLevel())
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return logger.Sugar()
}

func getJSONEncoder() zapcore.Encoder {
	// 自定义时间输出格式
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}

	// 自定义日志级别显示
	customLevelEncoder := func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(level.CapitalString())
	}

	// 定义zap配置信息
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     customTimeEncoder,          // 自定义时间格式
		EncodeLevel:    customLevelEncoder,         // 小写编码器
		EncodeCaller:   zapcore.ShortCallerEncoder, // 全路径编码器
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

type Fields map[string]interface{}

func getLoggerWrite() zapcore.WriteSyncer {

	// 定义日志切割配置
	l := &lumberjack.Logger{
		Filename:   path.Join("./logFile", "gb28181Panda.log"), //Filename 是要写入日志的文件。
		MaxSize:    1,                                          //MaxSize 是日志文件在轮换之前的最大大小（以兆字节为单位）。它默认为 100 兆字节
		MaxBackups: 30,                                         //MaxBackups 是要保留的最大旧日志文件数。默认是保留所有旧的日志文件（尽管 MaxAge 可能仍会导致它们被删除。）
		MaxAge:     30,                                         //MaxAge 是根据文件名中编码的时间戳保留旧日志文件的最大天数。
		Compress:   true,                                       //压缩
		LocalTime:  true,                                       //LocalTime 确定用于格式化备份文件中的时间戳的时间是否是计算机的本地时间。默认是使用 UTC 时间。
	}
	// 控制台输出
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(l))
}

func getLoggerLevel() zapcore.Level {
	level, _ := zapcore.ParseLevel("debug")
	return level
}

func Debug(args ...interface{}) {
	l.Debug(args)
}

func Debugf(format string, args ...interface{}) {
	l.Debugf(format, args)
}

func Info(args ...interface{}) {
	l.Info(args)
}

func Infof(format string, args ...interface{}) {
	l.Infof(format, args)
}

func Warn(args ...interface{}) {
	l.Warn(args)
}

func Warnf(format string, args ...interface{}) {
	l.Warnf(format, args)
}

func Error(args ...interface{}) {
	l.Error(args)
}

func Errorf(format string, args ...interface{}) {
	l.Errorf(format, args)
}

func Fatal(args ...interface{}) {
	l.Fatal(args)
}

func Fatalf(format string, args ...interface{}) {
	l.Fatalf(format, args)
}

func Panic(args ...interface{}) {
	l.Panic(args)
}

func Panicf(format string, args ...interface{}) {
	l.Panicf(format, args)
}
