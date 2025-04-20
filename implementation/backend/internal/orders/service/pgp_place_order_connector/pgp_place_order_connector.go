package pgp_place_order_connector

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"time"
	"tugas-akhir/backend/infrastructure/amqp"
	entity2 "tugas-akhir/backend/infrastructure/amqp/entity"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/pkg/logger"
	"tugas-akhir/backend/pkg/utility"
)

type PGPPlaceOrderConnector struct {
	placeOrderPublisher     *amqp.Publisher
	placeOrderReplyConsumer *amqp.Consumer
	ListenerChan            map[string]chan entity.PlaceOrderReplyMessage
	ctx                     context.Context
	config                  *config.Config
	ReplyRouteName          string
}

func NewPGPPlaceOrderConnector(
	ctx context.Context,
	config *config.Config,
) *PGPPlaceOrderConnector {
	placeOrderPublisher := amqp.NewPublisher(config, &entity.PlaceOrderExchange)

	replyRouteName := fmt.Sprintf("place_orders_reply.%s", config.PodName)

	placeOrderReplyConsumer := amqp.NewConsumer(
		config,
		entity.NewPlaceOrderReplyQueue(config.PodName),
		&entity.PlaceOrderExchange,
		entity2.ConsumeConfig{
			PrefetchCount: 10000,
			PrefetchSize:  0,
			AutoAck:       false,
			RoutingKeys:   []string{replyRouteName},
		},
	)

	return &PGPPlaceOrderConnector{
		ctx:                     ctx,
		config:                  config,
		placeOrderReplyConsumer: placeOrderReplyConsumer,
		placeOrderPublisher:     placeOrderPublisher,
		ReplyRouteName:          replyRouteName,
		ListenerChan:            make(map[string]chan entity.PlaceOrderReplyMessage),
	}
}

func (c *PGPPlaceOrderConnector) Stop() error {
	return c.placeOrderPublisher.Close()
}

func (c *PGPPlaceOrderConnector) Run() error {
	// run consume place order
	err := c.consumeReply()

	if err != nil {
		return err
	}

	return nil
}

func (c *PGPPlaceOrderConnector) consumeReply() error {
	l := logger.FromCtx(c.ctx).With(zap.String("context", "place-order-reply-consumer"))

	go func() {
	connLoop:
		for {
			channel, consumeErr := c.placeOrderReplyConsumer.Consume(c.ctx)

			if consumeErr != nil {
				l.Sugar().Error(consumeErr)
				continue
			}

		mainLoop:
			for {
				select {
				case rawMsg, ok := <-channel:
					if !ok {
						l.Info("channel is closed")
						break mainLoop
					}

					l.Info("receiving message")

					go func() {
						var buffer bytes.Buffer

						if _, writeBufferErr := buffer.Write(rawMsg.Body); writeBufferErr != nil {
							l.Sugar().Error(writeBufferErr)

							rejectErr := rawMsg.Reject(true)

							if rejectErr != nil {
								l.Sugar().Error(rejectErr)
							}

							return
						}

						var payload entity.PlaceOrderReplyMessage

						decoder := gob.NewDecoder(&buffer)
						if decodeErr := decoder.Decode(&payload); decodeErr != nil {
							l.Sugar().Error(decodeErr)

							rejectErr := rawMsg.Reject(true)

							if rejectErr != nil {
								l.Sugar().Error(rejectErr)

							}

							return
						}

						// pass the result to channel
						ch, exists := c.ListenerChan[payload.IdempotencyKey]

						if exists {
							ch <- payload
						} else {
							l.Warn("cannot find the corresponding listener", zap.String("idempotency-key", payload.IdempotencyKey))
						}
					}()
				case <-c.ctx.Done():
					l.Info("consume place order reply context is done")
					break connLoop
				}
			}
		}
	}()

	return nil
}

func (c *PGPPlaceOrderConnector) PublishRequest(ctx context.Context, message entity.PlaceOrderMessage) error {
	amqpMessage, err := message.ToMessage()

	if err != nil {
		return err
	}

ensureDelivered:
	for {
		err := c.placeOrderPublisher.Push(*amqpMessage)

		if err != nil {
			if errors.Is(err, entity2.NotConnectedError) {
				utility.SleepWithContext(ctx, 100*time.Millisecond)
			} else {
				return err
			}
		}

		break ensureDelivered
	}

	return nil
}
