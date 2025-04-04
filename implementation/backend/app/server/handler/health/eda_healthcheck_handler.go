package health

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"sync"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/infrastructure/redis"
	"tugas-akhir/backend/infrastructure/risingwave"
	myerror "tugas-akhir/backend/pkg/error"
	"tugas-akhir/backend/pkg/logger"
)

type EDAHealthcheckHandler struct {
	db    *postgres.Postgres
	rw    *risingwave.Risingwave
	redis *redis.Redis
}

func NewEDAHealthcheckHandler(
	db *postgres.Postgres,
	rw *risingwave.Risingwave,
	redis *redis.Redis,
) *EDAHealthcheckHandler {
	return &EDAHealthcheckHandler{
		db:    db,
		rw:    rw,
		redis: redis,
	}
}

type EDAHealthcheckResult struct {
	RedpandaStatus   string `json:"redpandaStatus"`
	RedisStatus      string `json:"redisStatus"`
	RisingwaveStatus string `json:"risingwaveStatus"`
}

func (h *EDAHealthcheckHandler) Healthcheck(c echo.Context) error {
	ctx := c.Request().Context()

	l := logger.FromCtx(ctx)

	status := http.StatusOK

	response := EDAHealthcheckResult{
		RedpandaStatus:   "Healthy",
		RedisStatus:      "Healthy",
		RisingwaveStatus: "Healthy",
	}

	wg := sync.WaitGroup{}

	wg.Add(2)

	go func() {
		err := h.rw.Pool.Ping(ctx)

		if err != nil {
			l.Sugar().Error(err)
			status = http.StatusServiceUnavailable
			response.RisingwaveStatus = err.Error()
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

	// todo healthcheck for redpanda

	wg.Wait()

	return c.JSON(status, myerror.HttpPayload{
		Message: "Ok",
		Data:    response,
	})
}
