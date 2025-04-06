package place_order

import (
	"context"
	errors2 "github.com/pkg/errors"
	"net/http"
	"time"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/orders/service/early_dropper"
	"tugas-akhir/backend/internal/orders/service/pgp_place_order_connector"
	myerror "tugas-akhir/backend/pkg/error"
	"tugas-akhir/backend/pkg/logger"
)

type FCPlaceOrderUsecase struct {
	config    *config.Config
	connector *pgp_place_order_connector.PGPPlaceOrderConnector
	dropper   *early_dropper.EarlyDropper
}

func NewFCPlaceOrderUsecase(
	config *config.Config,
	connector *pgp_place_order_connector.PGPPlaceOrderConnector,
	dropper *early_dropper.EarlyDropper,
) *FCPlaceOrderUsecase {

	return &FCPlaceOrderUsecase{
		connector: connector,
		config:    config,
		dropper:   dropper,
	}
}

func (u *FCPlaceOrderUsecase) PlaceOrder(ctx context.Context, payload entity.PlaceOrderDto) (*entity.Order, *myerror.HttpError) {
	lock, err := u.dropper.TryAcquireLock(ctx, payload)
	l := logger.FromCtx(ctx)

	if err != nil {
		if errors2.Is(err, entity.DropperSeatNotAvailable) {
			return nil, &myerror.HttpError{
				Code:    http.StatusConflict,
				Message: err.Error(),
			}
		} else if errors2.Is(err, entity.CannotAcquireLock) {
			return nil, &myerror.HttpError{
				Code:    http.StatusLocked,
				Message: err.Error(),
			}
		} else {
			return nil, &myerror.HttpError{
				Code:         http.StatusInternalServerError,
				Message:      err.Error(),
				ErrorContext: err,
			}
		}
	}

	// publish
	requestAmqpMessage := entity.PlaceOrderMessage{
		Data:       payload,
		ReplyRoute: u.connector.ReplyRouteName,
	}

	replyChan := make(chan entity.PlaceOrderReplyMessage, 1)
	u.connector.ListenerChan[*payload.IdempotencyKey] = replyChan

	defer func() {
		delete(u.connector.ListenerChan, *payload.IdempotencyKey)
		close(replyChan)
	}()

	if err := u.connector.PublishRequest(ctx, requestAmqpMessage); err != nil {

		if releaseErr := lock.OnFailed(); releaseErr != nil {
			l.Sugar().Error(releaseErr)
		}

		return nil, &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	select {
	case reply := <-replyChan:
		if reply.HttpErr == nil {
			if releaseErr := lock.OnSuccess(); releaseErr != nil {
				l.Sugar().Error(releaseErr)
			}
		} else {
			if releaseErr := lock.OnFailed(); releaseErr != nil {
				l.Sugar().Error(releaseErr)
			}
		}

		return reply.Order, reply.HttpErr
	case <-time.After(entity.PlaceOrderTimeout + 2*time.Second):
		if releaseErr := lock.OnFailed(); releaseErr != nil {
			l.Sugar().Error(releaseErr)
		}
		return nil, &myerror.HttpError{
			Code:    http.StatusRequestTimeout,
			Message: entity.PlaceOrderTimeoutError.Error(),
		}
	case <-ctx.Done():
		if releaseErr := lock.OnFailed(); releaseErr != nil {
			l.Sugar().Error(releaseErr)
		}
		return nil, &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      entity.PlaceOrderCancelled.Error(),
			ErrorContext: entity.PlaceOrderCancelled,
		}
	}
}
