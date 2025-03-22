package middleware

import (
	"fmt"
	"time"
	"tugas-akhir/backend/pkg/logger"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger is a middleware and zap to provide an "access log" like logging for each request.
func ZapLogger(l *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			beforeReq := c.Request()

			id := beforeReq.Header.Get(echo.HeaderXRequestID)

			ctx := c.Request().Context()

			log := l.With(zap.String("request_id", id))

			ctx = logger.WithCtx(ctx, log)

			beforeReq = c.Request().WithContext(ctx)

			c.SetRequest(beforeReq)

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			afterReq := c.Request()
			afterRes := c.Response()

			// take from
			log = logger.FromCtx(afterReq.Context())

			fields := []zapcore.Field{
				zap.String("remote_ip", c.RealIP()),
				zap.String("latency", time.Since(start).String()),
				zap.String("host", afterReq.Host),
				zap.String("request", fmt.Sprintf("%s %s", afterReq.Method, afterReq.RequestURI)),
				zap.Int("status", afterRes.Status),
				zap.Int64("size", afterRes.Size),
				zap.String("user_agent", afterReq.UserAgent()),
			}

			n := afterRes.Status
			switch {
			case n >= 500:
				log.With(zap.Error(err)).Error("Server error", fields...)
			case n >= 400:
				log.With(zap.Error(err)).Warn("Client error", fields...)
			case n >= 300:
				log.Info("Redirection", fields...)
			default:
				log.Info("Success", fields...)
			}

			return nil
		}
	}
}
