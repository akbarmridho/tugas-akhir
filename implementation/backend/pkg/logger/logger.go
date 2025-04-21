package logger

import (
	"context"
	"fmt"
	prettyconsole "github.com/thessem/zap-prettyconsole"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"sync"
)

type ctxKey struct{}

var once sync.Once

var logger *zap.Logger

var infoOnce sync.Once

var infoLogger *zap.Logger

func Get() *zap.Logger {
	once.Do(func() {
		isProduction := os.Getenv("ENVIRONMENT") == "production"

		stdout := zapcore.AddSync(os.Stdout)

		var level zapcore.Level

		levelEnv := os.Getenv("LOG_LEVEL")

		if levelEnv != "" {
			levelFromEnv, err := zapcore.ParseLevel(levelEnv)
			if err != nil {
				log.Println(
					fmt.Errorf("invalid level, defaulting to INFO: %w", err),
				)
			}

			level = levelFromEnv
		} else {
			if isProduction {
				level = zap.WarnLevel
			} else {
				level = zap.DebugLevel
			}
		}

		logLevel := zap.NewAtomicLevelAt(level)

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

		logger = zap.New(core)
		defer logger.Sync()
	})

	return logger
}

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
	} else if l := logger; l != nil {
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
