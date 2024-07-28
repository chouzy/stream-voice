package logger

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"strings"
	"time"
)

var zapConf *Zap

type Zap struct {
	Level         string `mapstructure:"level" json:"level" yaml:"level"`                         // 级别
	Prefix        string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`                      // 日志前缀
	Format        string `mapstructure:"format" json:"format" yaml:"format"`                      // 输出
	Director      string `mapstructure:"director" json:"director"  yaml:"director"`               // 日志文件夹
	EncodeLevel   string `mapstructure:"encodeLevel" json:"encodeLevel" yaml:"encodeLevel"`       // 编码级别
	StacktraceKey string `mapstructure:"stacktraceKey" json:"stacktraceKey" yaml:"stacktraceKey"` // 栈名

	MaxAge       int  `mapstructure:"maxAge" json:"maxAge" yaml:"maxAge"`                   // 日志留存时间
	ShowLine     bool `mapstructure:"showLine" json:"showLine" yaml:"showLine"`             // 显示行
	LogInConsole bool `mapstructure:"logInConsole" json:"logInConsole" yaml:"logInConsole"` // 输出控制台
}

// ZapEncodeLevel 获取编码器级别
func (z *Zap) ZapEncodeLevel() zapcore.LevelEncoder {
	// 以json格式输出时如果使用带颜色的编码器会出现乱码
	switch z.EncodeLevel {
	case "LowercaseLevelEncoder": // 小写编码器
		return zapcore.LowercaseLevelEncoder
	case "LowercaseColorLevelEncoder": // 小写编码器带颜色
		return zapcore.LowercaseColorLevelEncoder
	case "CapitalLevelEncoder": // 大写编码器
		return zapcore.CapitalLevelEncoder
	case "CapitalColorLevelEncoder": // 大写编码器带颜色
		return zapcore.CapitalColorLevelEncoder
	default:
		return zapcore.LowercaseLevelEncoder
	}
}

// TransportLevel 将日志级别转换为 zapcore.Level 类型
func (z *Zap) TransportLevel() zapcore.Level {
	z.Level = strings.ToLower(z.Level)
	switch z.Level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.WarnLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.DebugLevel
	}
}

// InitZap 初始化日志设置
func InitZap(z *Zap) *zap.Logger {
	zapConf = z
	cores := getZapCores()
	logger := zap.New(zapcore.NewTee(cores...))
	if zapConf.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}
	return logger
}

// getZapCores 日志输出核心
func getZapCores() []zapcore.Core {
	cores := make([]zapcore.Core, 0, 7)
	for level := zapConf.TransportLevel(); level <= zapcore.FatalLevel; level++ {
		cores = append(cores, getEncodeCore(level, getLevelPriority(level)))
	}
	return cores
}

// getLevelPriority 获取优先级
func getLevelPriority(level zapcore.Level) zap.LevelEnablerFunc {
	switch level {
	case zapcore.DebugLevel:
		return func(level zapcore.Level) bool { // 调试级别
			return level == zap.DebugLevel
		}
	case zapcore.InfoLevel:
		return func(level zapcore.Level) bool { // 日志级别
			return level == zap.InfoLevel
		}
	case zapcore.WarnLevel:
		return func(level zapcore.Level) bool { // 警告级别
			return level == zap.WarnLevel
		}
	case zapcore.ErrorLevel:
		return func(level zapcore.Level) bool { // 错误级别
			return level == zap.ErrorLevel
		}
	case zapcore.DPanicLevel:
		return func(level zapcore.Level) bool { // dpanic级别
			return level == zap.DPanicLevel
		}
	case zapcore.PanicLevel:
		return func(level zapcore.Level) bool { // panic级别
			return level == zap.PanicLevel
		}
	case zapcore.FatalLevel:
		return func(level zapcore.Level) bool { // 终止级别
			return level == zap.FatalLevel
		}
	default:
		return func(level zapcore.Level) bool { // 调试级别
			return level == zap.DebugLevel
		}
	}
}

// getEncodeCore 获取EncodeCore
func getEncodeCore(l zapcore.Level, level zap.LevelEnablerFunc) zapcore.Core {
	write, err := getWriteSyncer(l.String())
	if err != nil {
		fmt.Printf("Get write syncer failed error: %v \n", err)
		return nil
	}
	return zapcore.NewCore(getEncoder(), write, level)
}

// getWriteSyncer 日志输出方式
func getWriteSyncer(level string) (zapcore.WriteSyncer, error) {
	fileWriter, err := rotatelogs.New(
		path.Join(zapConf.Director, "%Y-%m-%d-"+level+".log"),
		rotatelogs.WithClock(rotatelogs.Local),
		rotatelogs.WithMaxAge(time.Duration(zapConf.MaxAge)*24*time.Hour),
		rotatelogs.WithRotationTime(time.Hour*24),
	)
	if zapConf.LogInConsole {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
	}
	return zapcore.AddSync(fileWriter), err
}

// getEncoder 获取日志编码器
func getEncoder() zapcore.Encoder {
	if zapConf.Format == "json" {
		return zapcore.NewJSONEncoder(getEncoderConfig())
	}
	return zapcore.NewConsoleEncoder(getEncoderConfig())
}

// getEncoderConfig 日志编码格式
func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "log",
		CallerKey:      "caller",
		StacktraceKey:  zapConf.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapConf.ZapEncodeLevel(),
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
}

func customTimeEncoder(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
	encoder.AppendString(t.Format("2006/01/02 15:04:05.000"))
}
