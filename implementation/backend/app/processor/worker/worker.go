package worker

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/orders/usecase/place_order"
	"tugas-akhir/backend/pkg/logger"
)

type BookingWorker struct {
	placeOrderUsecase place_order.PlaceOrderUsecase
	resultPublisher   *ResultPublisher
}

func (w *BookingWorker) Process(ctx context.Context, rawMsg *amqp091.Delivery) error {
	l := logger.FromCtx(ctx).With(zap.String("context", "worker"))

	var buffer bytes.Buffer

	if _, writeBufferErr := buffer.Write(rawMsg.Body); writeBufferErr != nil {
		l.Sugar().Error(writeBufferErr)

		rejectErr := rawMsg.Reject(true)

		if rejectErr != nil {
			l.Sugar().Error(rejectErr)
			return rejectErr
		}

		return writeBufferErr
	}

	var payload entity.PlaceOrderMessage

	decoder := gob.NewDecoder(&buffer)
	if decodeErr := decoder.Decode(&payload); decodeErr != nil {
		l.Sugar().Error(decodeErr)

		rejectErr := rawMsg.Reject(true)

		if rejectErr != nil {
			l.Sugar().Error(rejectErr)

		}

		return decodeErr
	}

	response, httpErr := w.placeOrderUsecase.PlaceOrder(ctx, payload.Data)

	publishErr := w.resultPublisher.Publish(ctx, entity.PlaceOrderReplyMessage{
		Order:      response,
		HttpErr:    httpErr,
		ReplyRoute: payload.ReplyRoute,
	})

	if publishErr != nil {
		l.Sugar().Error(publishErr)
		return publishErr
	}

	return nil
}

func (w *BookingWorker) Stop() error {
	return w.resultPublisher.Stop()
}

func NewBookingWorker(
	placeOrderUsecase place_order.PlaceOrderUsecase,
	resultPublisher *ResultPublisher,
) *BookingWorker {
	return &BookingWorker{
		placeOrderUsecase: placeOrderUsecase,
		resultPublisher:   resultPublisher,
	}
}
