package place_order

import (
	"context"
	"net/http"
	"time"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/orders/service/pgp_place_order_connector"
	myerror "tugas-akhir/backend/pkg/error"
)

type PGPPlaceOrderUsecase struct {
	config    *config.Config
	connector *pgp_place_order_connector.PGPPlaceOrderConnector
}

func NewPGPPlaceOrderUsecase(
	config *config.Config,
	connector *pgp_place_order_connector.PGPPlaceOrderConnector,
) *PGPPlaceOrderUsecase {

	return &PGPPlaceOrderUsecase{
		connector: connector,
		config:    config,
	}
}

func (u *PGPPlaceOrderUsecase) PlaceOrder(ctx context.Context, payload entity.PlaceOrderDto) (*entity.Order, *myerror.HttpError) {
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
		return nil, &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	select {
	case reply := <-replyChan:
		return reply.Order, reply.HttpErr
	case <-time.After(entity.PlaceOrderTimeout + 2*time.Second):
		return nil, &myerror.HttpError{
			Code:    http.StatusRequestTimeout,
			Message: entity.PlaceOrderTimeoutError.Error(),
		}
	case <-ctx.Done():
		return nil, &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      entity.PlaceOrderCancelled.Error(),
			ErrorContext: entity.PlaceOrderCancelled,
		}
	}
}
