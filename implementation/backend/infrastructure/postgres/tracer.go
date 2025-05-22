package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
)

type ZapTracer struct {
	logger *zap.Logger
}

func NewZapTracer(logger *zap.Logger) *ZapTracer {
	return &ZapTracer{logger: logger}
}

func (z *ZapTracer) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	fields := make([]zap.Field, 0, len(data))
	for k, v := range data {
		fields = append(fields, zap.Any(k, v))
	}

	switch level {
	case tracelog.LogLevelDebug:
		z.logger.Debug(msg, fields...)
	case tracelog.LogLevelInfo:
		z.logger.Info(msg, fields...)
	case tracelog.LogLevelWarn:
		z.logger.Warn(msg, fields...)
	case tracelog.LogLevelError:
		z.logger.Error(msg, fields...)
	default:
		z.logger.Info(msg, fields...) // fallback to Info
	}
}
