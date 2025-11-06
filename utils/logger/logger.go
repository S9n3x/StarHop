package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

func init() {
	config := zap.NewProductionConfig()
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config.EncoderConfig),
		zapcore.Lock(os.Stdout),
		zapcore.DebugLevel,
	)

	zapLogger := zap.New(core)
	logger = zapLogger.Sugar()
}

func Debug(args ...interface{}) {
	logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Debugw(fmt.Sprint(args...))
}

// Info
func Info(args ...interface{}) {
	logger.Infow(fmt.Sprint(args...))
}

// Warn
func Warn(args ...interface{}) {
	logger.Warnw(fmt.Sprint(args...))
}

// Error
func Error(args ...interface{}) {
	logger.Errorw(fmt.Sprint(args...))
	logger.Sync()
	os.Exit(1)
}
