package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"strings"
	"time"
	"tugas-akhir/backend/internal/orders/service/early_dropper"
	"tugas-akhir/backend/pkg/logger"
)

type TimeoutMiddleware struct {
	TimeoutMiddleware echo.MiddlewareFunc
}

func NewTimeoutMiddleware() *TimeoutMiddleware {
	return &TimeoutMiddleware{
		TimeoutMiddleware: middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			Timeout: 3 * time.Minute,
			OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
				if c.Request().Method == "POST" && strings.Contains(c.Request().URL.Path, "orders") {
					l := logger.GetInfo()

					requestID := c.Response().Header().Get(echo.HeaderXRequestID)

					if requestID == "" {
						requestID = c.Request().Header.Get(echo.HeaderXRequestID)
					}

					if requestID == "" {
						l.Warn("Cannot find request id. skipping timeout callback check")
						return
					}

					value, ok := early_dropper.LockHolder.Locks.Load(requestID)

					if !ok {
						l.Warn("Cannot find lock releaser. skipping")
						return
					}

					lockReleaser, ok := value.(*early_dropper.LockReleaser)

					if ok {
						releaseErr := lockReleaser.OnFailed()

						if releaseErr != nil {
							l.Error("lock release failed", zap.Error(releaseErr))
						}
					}
				}
			},
		}),
	}
}
