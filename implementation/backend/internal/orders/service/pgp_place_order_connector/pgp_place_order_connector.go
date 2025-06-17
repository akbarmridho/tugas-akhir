package pgp_place_order_connector

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"sync"
	"time"
	"tugas-akhir/backend/infrastructure/amqp"
	entity2 "tugas-akhir/backend/infrastructure/amqp/entity"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/pkg/logger"
	"tugas-akhir/backend/pkg/utility"
)

type FCPlaceOrderConnector struct {
	placeOrderPublisher     *amqp.Publisher
	placeOrderReplyConsumer *amqp.Consumer
	//ListenerChan            map[string]chan entity.PlaceOrderReplyMessage
	ListenerChan   sync.Map
	ctx            context.Context
	cancelCtx      context.CancelFunc
	config         *config.Config
	ReplyRouteName string
}

func NewFCPlaceOrderConnector(
	config *config.Config,
) *FCPlaceOrderConnector {
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

	ctx, cancel := context.WithCancel(context.Background())

	return &FCPlaceOrderConnector{
		ctx:                     ctx,
		cancelCtx:               cancel,
		config:                  config,
		placeOrderReplyConsumer: placeOrderReplyConsumer,
		placeOrderPublisher:     placeOrderPublisher,
		ReplyRouteName:          replyRouteName,
		ListenerChan:            sync.Map{},
	}
}

func (c *FCPlaceOrderConnector) Stop() error {
	c.cancelCtx()

	err1 := c.placeOrderPublisher.Close()
	err2 := c.placeOrderReplyConsumer.Close()

	if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	}

	return nil
}

func (c *FCPlaceOrderConnector) Run(ctx context.Context) error {
	// run consume place order
	err := c.consumeReply()

	if err != nil {
		return err
	}

	return nil
}

func (c *FCPlaceOrderConnector) consumeReply() error {
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

					//l.Info("receiving message")

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
						//ch, exists := c.ListenerChan[payload.IdempotencyKey]
						rawCh, exists := c.ListenerChan.Load(payload.IdempotencyKey)

						if exists {
							ch := rawCh.(chan entity.PlaceOrderReplyMessage)

							ch <- payload

							if ackErr := rawMsg.Ack(false); ackErr != nil {
								l.Error("ack err", zap.Error(ackErr))
							}

						} else {
							l.Warn("cannot find the corresponding listener", zap.String("idempotency-key", payload.IdempotencyKey))
							// just ack in this case to prevent redelivery
							if ackErr := rawMsg.Ack(false); ackErr != nil {
								l.Error("ack err", zap.Error(ackErr))
							}
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

func (c *FCPlaceOrderConnector) PublishRequest(ctx context.Context, message entity.PlaceOrderMessage) error {
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
