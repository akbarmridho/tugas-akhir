package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	redis2 "github.com/redis/go-redis/v9"
	"net/http"
	"time"
	"tugas-akhir/backend/infrastructure/redis"
	"tugas-akhir/backend/internal/orders/entity"
	myerror "tugas-akhir/backend/pkg/error"
	"tugas-akhir/backend/pkg/logger"
)

func buildRedisIdempotencyKey(key string) string {
	return fmt.Sprintf("place-order:%s", key)
}

type idempotencyData struct {
	entity  *entity.Order
	httpErr *myerror.HttpError
}

func WrapIdempotency(
	ctx context.Context,
	redis *redis.Redis,
	handler func(ctx context.Context, payload entity.PlaceOrderDto) (*entity.Order, *myerror.HttpError),
	payload entity.PlaceOrderDto,
) (*entity.Order, *myerror.HttpError) {
	l := logger.FromCtx(ctx)

	if payload.IdempotencyKey == nil || *payload.IdempotencyKey != "" {
		return nil, &myerror.HttpError{
			Code:    http.StatusBadRequest,
			Message: entity.IdempotencyKeyNotFound.Error(),
		}
	}

	cacheVal, cacheErr := redis.Client.Get(ctx, buildRedisIdempotencyKey(*payload.IdempotencyKey)).Result()

	if cacheErr != nil {
		if errors.Is(cacheErr, redis2.Nil) {
			// perform the request as usual
			order, httpErr := handler(ctx, payload)

			if httpErr == nil {
				// store idempotency for success operation only
				cacheData := idempotencyData{
					entity:  order,
					httpErr: httpErr,
				}

				marshalled, err := json.Marshal(cacheData)

				if err != nil {
					l.Sugar().Error(err)
				}

				if cmdErr := redis.Client.SetEx(ctx, buildRedisIdempotencyKey(*payload.IdempotencyKey), marshalled, 15*time.Minute).Err(); cmdErr != nil {
					l.Sugar().Error(cmdErr)
				}
			}

			return order, httpErr
		} else {
			return nil, &myerror.HttpError{
				Code:         http.StatusInternalServerError,
				Message:      cacheErr.Error(),
				ErrorContext: cacheErr,
			}
		}
	}

	// value found
	var result idempotencyData
	if err := json.Unmarshal([]byte(cacheVal), &result); err != nil {
		return nil, &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	return result.entity, result.httpErr
}
