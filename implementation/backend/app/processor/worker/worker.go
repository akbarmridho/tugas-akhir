package worker

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"time"
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

	waitTime := time.Since(rawMsg.Timestamp)
	now := time.Now()

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

	processTime := time.Since(now)
	now = time.Now()

	publishErr := w.resultPublisher.Publish(ctx, entity.PlaceOrderReplyMessage{
		Order:          response,
		HttpErr:        httpErr,
		ReplyRoute:     payload.ReplyRoute,
		IdempotencyKey: *payload.Data.IdempotencyKey,
	})

	publishTime := time.Since(now)

	if publishErr != nil {
		l.Sugar().Error(publishErr)
		return publishErr
	}

	l.Info("order processed",
		zap.Int64("wait_time", waitTime.Milliseconds()),
		zap.Int64("process_time", processTime.Milliseconds()),
		zap.Int64("publish_time", publishTime.Milliseconds()),
	)

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
