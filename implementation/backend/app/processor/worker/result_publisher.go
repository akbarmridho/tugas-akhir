package worker

import (
	"context"
	"errors"
	"time"
	"tugas-akhir/backend/infrastructure/amqp"
	entity2 "tugas-akhir/backend/infrastructure/amqp/entity"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/pkg/utility"
)

type ResultPublisher struct {
	replyPublisher *amqp.Publisher
}

func (p *ResultPublisher) Stop() error {
	return p.replyPublisher.Close()
}

func (p *ResultPublisher) Publish(ctx context.Context, message entity.PlaceOrderReplyMessage) error {
	amqpMessage, err := message.ToMessage()

	if err != nil {
		return err
	}

ensureDelivered:
	for {
		err := p.replyPublisher.Push(*amqpMessage)

		if err != nil {
			if errors.Is(err, entity2.NotConnectedError) {
				utility.SleepWithContext(ctx, 2*time.Second)
			} else {
				return err
			}
		}

		break ensureDelivered
	}

	return nil
}

func NewResultPublisher(config *config.Config) *ResultPublisher {
	replyPublisher := amqp.NewPublisher(config, &entity.PlaceOrderExchange)

	return &ResultPublisher{
		replyPublisher: replyPublisher,
	}
}
