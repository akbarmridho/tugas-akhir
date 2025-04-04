package health

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"sync"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/infrastructure/redis"
	myerror "tugas-akhir/backend/pkg/error"
	"tugas-akhir/backend/pkg/logger"
)

type BaseHealthcheckHandler struct {
	db    *postgres.Postgres
	redis *redis.Redis
}

func NewBaseHealthcheckHandler(db *postgres.Postgres, redis *redis.Redis) *BaseHealthcheckHandler {
	return &BaseHealthcheckHandler{
		db:    db,
		redis: redis,
	}
}

type PGHealthcheckResult struct {
	PostgresStatus string `json:"postgresStatus"`
	RedisStatus    string `json:"redisStatus"`
}

func (h *BaseHealthcheckHandler) Healthcheck(c echo.Context) error {
	ctx := c.Request().Context()

	l := logger.FromCtx(ctx)

	status := http.StatusOK

	response := PGHealthcheckResult{
		PostgresStatus: "Healthy",
		RedisStatus:    "Healthy",
	}

	wg := sync.WaitGroup{}

	wg.Add(2)

	go func() {
		err := h.db.Pool.Ping(ctx)

		if err != nil {
			l.Sugar().Error(err)
			status = http.StatusServiceUnavailable
			response.PostgresStatus = err.Error()
		}

		wg.Done()
	}()

	go func() {
		err := h.redis.IsHealthy(ctx)

		if err != nil {
			l.Sugar().Error(err)
			status = http.StatusServiceUnavailable
			response.RedisStatus = err.Error()
		}

		wg.Done()
	}()

	wg.Wait()

	return c.JSON(status, myerror.HttpPayload{
		Message: "Ok",
		Data:    response,
	})
}
