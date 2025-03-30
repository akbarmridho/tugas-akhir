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
	ctx               context.Context
	placeOrderUsecase place_order.PlaceOrderUsecase
	resultPublisher   *ResultPublisher
}

// Process will synchronously perform a job and return the result.
func (w *BookingWorker) Process(rawPayload interface{}) interface{} {
	l := logger.FromCtx(w.ctx).With(zap.String("context", "worker"))

	rawMsg := rawPayload.(amqp091.Delivery)

	var buffer bytes.Buffer

	_, writeBufferErr := buffer.Write(rawMsg.Body)

	if writeBufferErr != nil {
		l.Sugar().Error(writeBufferErr)

		rejectErr := rawMsg.Reject(true)

		if rejectErr != nil {
			l.Sugar().Error(rejectErr)

		}

		return nil
	}

	var payload entity.PlaceOrderMessage

	decoder := gob.NewDecoder(&buffer)
	decodeErr := decoder.Decode(&payload)

	if decodeErr != nil {
		l.Sugar().Error(decodeErr)

		rejectErr := rawMsg.Reject(true)

		if rejectErr != nil {
			l.Sugar().Error(rejectErr)

		}

		return nil
	}

	response, httpErr := w.placeOrderUsecase.PlaceOrder(w.ctx, payload.Data)

	return w.resultPublisher.Publish(w.ctx, entity.PlaceOrderReplyMessage{
		Order:      response,
		HttpErr:    httpErr,
		ReplyRoute: payload.ReplyRoute,
	})
}

// BlockUntilReady is called before each job is processed and must block the
// calling goroutine until the Worker is ready to process the next job.
func (w *BookingWorker) BlockUntilReady() {
	// todo
}

// Interrupt is called when a job is cancelled. The worker is responsible
// for unblocking the Process implementation.
func (w *BookingWorker) Interrupt() {
	// todo
}

// Terminate is called when a Worker is removed from the processing pool
// and is responsible for cleaning up any held resources.
func (w *BookingWorker) Terminate() {
	// todo
}

func NewBookingWorker(
	ctx context.Context,
	placeOrderUsecase place_order.PlaceOrderUsecase,
	resultPublisher *ResultPublisher,
) *BookingWorker {
	return &BookingWorker{
		ctx:               ctx,
		placeOrderUsecase: placeOrderUsecase,
		resultPublisher:   resultPublisher,
	}
}
