package health

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"sync"
	"tugas-akhir/backend/infrastructure/amqp"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/infrastructure/redis"
	myerror "tugas-akhir/backend/pkg/error"
	"tugas-akhir/backend/pkg/logger"
)

type FCHealthcheckHandler struct {
	db    *postgres.Postgres
	redis *redis.Redis
}

func NewFCHealthcheckHandler(db *postgres.Postgres, redis *redis.Redis) *FCHealthcheckHandler {
	return &FCHealthcheckHandler{
		db:    db,
		redis: redis,
	}
}

type FCHealthcheckResult struct {
	PostgresStatus string `json:"postgresStatus"`
	RedisStatus    string `json:"redisStatus"`
	AmqpStatus     string `json:"amqpStatus"`
}

func (h *FCHealthcheckHandler) Healthcheck(c echo.Context) error {
	ctx := c.Request().Context()

	l := logger.FromCtx(ctx)

	status := http.StatusOK

	response := FCHealthcheckResult{
		PostgresStatus: "Healthy",
		RedisStatus:    "Healthy",
		AmqpStatus:     "Healthy",
	}

	wg := sync.WaitGroup{}

	wg.Add(3)

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

	go func() {
		everyConnected := true

		for _, consumer := range amqp.ConnectedConsumers {
			if consumer == nil || !consumer.Client.IsConnected() {
				everyConnected = false
				break
			}
		}

		for _, publisher := range amqp.ConnectedPublishers {
			if publisher == nil || !publisher.Client.IsConnected() {
				everyConnected = false
				break
			}
		}

		if !everyConnected {
			theErr := fmt.Errorf("some RabbitMQ publisher or consumer is not connected")
			l.Sugar().Error(theErr)
			status = http.StatusServiceUnavailable
			response.AmqpStatus = theErr.Error()
		}

		wg.Done()
	}()

	wg.Wait()

	return c.JSON(status, myerror.HttpPayload{
		Message: "Ok",
		Data:    response,
	})
}
