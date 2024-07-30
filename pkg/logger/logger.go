package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Fields map[string]interface{}

type Logger struct {
	newLogger *zap.Logger
	fields    Fields
	callers   []string
}

func NewLogger(hook io.Writer, mode string) *Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "tag",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime: func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(time.Format("2006-01-02 15:04:05"))
		}, // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.InfoLevel)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 编码器配置
		// zapcore.NewMultiWriteSyncer(zapcore.AddSync(hook)), //打印到文件 --打印到控制台：zapcore.AddSync(os.Stdout)
		zapcore.AddSync(os.Stdout),
		atomicLevel, // 日志级别
	)
	// 输出文件和行号，前提是配置对象encoderConfig中必须设有CallerKey字段
	caller := zap.AddCaller()
	// 由于再次封装日志，因此需要打印上一级的调用，1表示向上跳一级
	callerSkip := zap.AddCallerSkip(1)

	if mode == "debug" {
		// 开启开发模式
		return &Logger{
			newLogger: zap.New(core, caller, callerSkip, zap.Development()),
		}
	}

	return &Logger{
		newLogger: zap.New(core, caller, callerSkip),
	}
}

func (l *Logger) clone() *Logger { // 防止并发时的数据脏乱
	nl := *l
	return &nl
}

func (l *Logger) WithFields(f Fields) *Logger {
	ll := l.clone()
	if ll.fields == nil {
		ll.fields = make(Fields)
	}
	for k, v := range f {
		ll.fields[k] = v
	}
	return ll
}

func (l *Logger) WithCallersFrames() *Logger {
	maxCallerDepth := 25
	minCallerDepth := 1
	callers := []string{}
	pcs := make([]uintptr, maxCallerDepth)
	depth := runtime.Callers(minCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		callers = append(callers, fmt.Sprintf("%s: %d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	ll := l.clone()
	ll.callers = callers
	return ll
}

func (l *Logger) Debug(msg string) {
	if l.fields != nil {
		l.newLogger.Debug(msg, zap.Any("field", l.fields))
		return
	}
	ll := l.clone()
	ll.newLogger.Debug(msg)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if l.fields != nil {
		l.newLogger.Debug(msg, zap.Any("field", l.fields))
		return
	}
	ll := l.clone()
	ll.newLogger.Debug(msg)
}

func (l *Logger) Info(msg string) {
	if l.fields != nil {
		l.newLogger.Info(msg, zap.Any("field", l.fields))
		return
	}
	ll := l.clone()
	ll.newLogger.Info(msg)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if l.fields != nil {
		l.newLogger.Info(msg, zap.Any("field", l.fields))
		return
	}
	ll := l.clone()
	ll.newLogger.Info(msg)
}

func (l *Logger) Warn(msg string) {
	if l.fields != nil {
		l.newLogger.Warn(msg, zap.Any("field", l.fields))
		return
	}
	ll := l.clone()
	ll.newLogger.Warn(msg)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if l.fields != nil {
		l.newLogger.Warn(msg, zap.Any("field", l.fields))
		return
	}
	ll := l.clone()
	ll.newLogger.Warn(msg)
}

func (l *Logger) Error(msg string) {
	if l.fields != nil {
		l.newLogger.Error(msg, zap.Any("field", l.fields))
		return
	}
	ll := l.clone()
	ll.newLogger.Error(msg)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if l.fields != nil {
		l.newLogger.Error(msg, zap.Any("field", l.fields))
		return
	}
	ll := l.clone()
	ll.newLogger.Error(msg)
}

func (l *Logger) Fatal(msg string) {
	if l.fields != nil {
		l.newLogger.Fatal(msg, zap.Any("field", l.fields))
		return
	}
	ll := l.clone()
	ll.newLogger.Fatal(msg)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if l.fields != nil {
		l.newLogger.Fatal(msg, zap.Any("field", l.fields))
		return
	}
	ll := l.clone()
	ll.newLogger.Fatal(msg)
}

func (l *Logger) Panic(msg string) {
	if l.fields != nil {
		l.newLogger.Panic(msg, zap.Any("field", l.fields))
		return
	}
	ll := l.clone()
	ll.newLogger.Panic(msg)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if l.fields != nil {
		l.newLogger.Panic(msg, zap.Any("field", l.fields))
		return
	}
	ll := l.clone()
	ll.newLogger.Panic(msg)
}
