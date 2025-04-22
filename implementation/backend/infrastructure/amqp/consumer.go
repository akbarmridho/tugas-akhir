package amqp

import (
	"context"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"tugas-akhir/backend/infrastructure/amqp/entity"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/pkg/logger"

	"sync"
)

type Consumer struct {
	Client
	consumeConfig entity.ConsumeConfig
}

func NewConsumer(
	config *config.Config,
	queue entity.QueueConfig,
	exchange *entity.ExchangeConfig,
	consumerConfig entity.ConsumeConfig,
) *Consumer {
	l := logger.GetInfo().With(zap.String("service", "amqp.consumer"))

	consumer := Consumer{
		consumeConfig: consumerConfig,
		Client: Client{
			M:        &sync.Mutex{},
			Logger:   l,
			Queue:    &queue,
			Exchange: exchange,
			done:     make(chan bool),
		},
	}

	go consumer.handleReconnect(config.AmqpUrl)

	ConnectedConsumers = append(ConnectedConsumers, &consumer)

	return &consumer
}

// Consume will continuously put queue items on the channel.
// It is required to call delivery.Ack when it has been
// successfully processed, or delivery.Nack when it fails.
// Ignoring this will cause data to build up on the server.
// NOTE
// BE CAREFUL WHEN USING CONSUME AS IT DOES NOT HANDLE CHANNEL RECONNECTION
// IF THE CHANNEL THAT ARE USED TO CONSUME THE MESSAGE ARE CLOSED, A NEW CHANNEL WILL BE CREATED
// AND YOU HAVE TO MANUALLY RE-CONSUME THE CHANNEL
func (client *Consumer) Consume(ctx context.Context) (<-chan amqp091.Delivery, error) {
	//client.M.Lock()
	//if !client.isReady {
	//	client.M.Unlock()
	//	return nil, entity.NotConnectedError
	//}
	//client.M.Unlock()
	client.WaitUntilReady(ctx)

	if err := client.channel.Qos(
		client.consumeConfig.PrefetchCount, // prefetchCount
		client.consumeConfig.PrefetchSize,  // prefetchSize
		false,                              // global
	); err != nil {
		return nil, err
	}

	exchangeName := ""

	if client.Exchange != nil {
		exchangeName = client.Exchange.Name
	}

	if len(client.consumeConfig.RoutingKeys) == 0 {
		err := client.channel.QueueBind(
			client.Queue.Name,
			"",
			exchangeName,
			client.Queue.NoWait,
			nil,
		)

		if err != nil {
			return nil, err
		}
	} else {
		for _, route := range client.consumeConfig.RoutingKeys {
			err := client.channel.QueueBind(
				client.Queue.Name,
				route,
				exchangeName,
				client.Queue.NoWait,
				nil,
			)

			if err != nil {
				return nil, err
			}
		}
	}

	return client.channel.Consume(
		client.Queue.Name,
		"",                           // Consumer
		client.consumeConfig.AutoAck, // Auto-Ack
		client.Queue.Exclusive,       // Exclusive
		false,                        // No-local
		client.Queue.NoWait,          // No-Wait
		nil,                          // Args
	)
}
