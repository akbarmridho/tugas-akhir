package amqp

import (
	"context"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
	"tugas-akhir/backend/infrastructure/amqp/entity"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/pkg/logger"
)

type Publisher struct {
	Client
}

func NewPublisher(
	config *config.Config,
	exchange *entity.ExchangeConfig,
) *Publisher {
	l := logger.GetInfo().With(zap.String("service", "amqp.publisher"))

	publisher := Publisher{
		Client: Client{
			M:        &sync.Mutex{},
			Logger:   l,
			Exchange: exchange,
			done:     make(chan bool),
		},
	}

	go publisher.handleReconnect(config.AmqpUrl)

	ConnectedPublishers = append(ConnectedPublishers, &publisher)

	return &publisher
}

// Push will push data onto the queue, and wait for a confirmation.
// This will block until the server sends a confirmation. Errors are
// only returned if the push action itself fails, see UnsafePush.
func (client *Publisher) Push(message entity.Message) error {
	client.M.Lock()
	if !client.isReady {
		client.M.Unlock()
		return entity.PushFailedNotConnectedError
	}
	client.M.Unlock()
	for {
		err := client.UnsafePush(message)
		if err != nil {
			client.Logger.Error("push failed. Retrying...")
			select {
			case <-client.done:
				return entity.ShutdownError
			case <-time.After(entity.ResendDelay):
			}
			continue
		}
		confirm := <-client.notifyConfirm
		if confirm.Ack {
			if message.LogDelivery {
				client.Logger.Sugar().Infof("push confirmed [%d]", confirm.DeliveryTag)
			}
			return nil
		}
	}
}

// UnsafePush will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// receive the message.
func (client *Publisher) UnsafePush(message entity.Message) error {
	client.M.Lock()
	if !client.isReady {
		client.M.Unlock()
		return entity.NotConnectedError
	}
	client.M.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exchangeName := ""

	if client.Exchange != nil {
		exchangeName = client.Exchange.Name
	}

	payload := amqp091.Publishing{
		ContentType: message.ContentType,
		Body:        message.Data,
	}

	if message.Type != nil {
		payload.Type = *message.Type
	}

	if message.TTL != nil {
		payload.Expiration = strconv.FormatInt(message.TTL.Milliseconds(), 10)
	}

	if message.IsPersistent {
		payload.DeliveryMode = 2
	}

	if message.Priority != nil {
		payload.Priority = *message.Priority
	}

	return client.channel.PublishWithContext(
		ctx,
		exchangeName,       // Exchange
		message.RoutingKey, // Routing key
		false,              // Mandatory
		false,              // Immediate
		payload,
	)
}
