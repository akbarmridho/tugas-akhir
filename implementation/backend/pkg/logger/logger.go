package logger

import (
	"context"
	prettyconsole "github.com/thessem/zap-prettyconsole"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
)

type ctxKey struct{}

var infoOnce sync.Once

var infoLogger *zap.Logger

func GetInfo() *zap.Logger {
	infoOnce.Do(func() {
		stdout := zapcore.AddSync(os.Stdout)

		level := zap.InfoLevel

		logLevel := zap.NewAtomicLevelAt(level)

		isProduction := os.Getenv("ENVIRONMENT") == "production"

		var encoder zapcore.Encoder

		if isProduction {
			encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		} else {
			//encoder = zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())
			encoder = prettyconsole.NewEncoder(prettyconsole.NewEncoderConfig())
		}

		core := zapcore.NewTee(
			zapcore.NewCore(encoder, stdout, logLevel),
		)

		infoLogger = zap.New(core)
		defer infoLogger.Sync()
	})

	return infoLogger
}

// FromCtx returns the Logger associated with the ctx. If no logger
// is associated, the default logger is returned, unless it is nil
// in which case a disabled logger is returned.
func FromCtx(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return l
	} else if l := infoLogger; l != nil {
		return l
	}

	return GetInfo()
}

// WithCtx returns a copy of ctx with the Logger attached.
func WithCtx(ctx context.Context, l *zap.Logger) context.Context {
	if lp, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		if lp == l {
			// Do not store same logger.
			return ctx
		}
	}

	return context.WithValue(ctx, ctxKey{}, l)
}
