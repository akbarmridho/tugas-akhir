package processor

import (
	"context"
	"github.com/platinummonkey/go-concurrency-limits/core"
	"github.com/platinummonkey/go-concurrency-limits/limit"
	limiter2 "github.com/platinummonkey/go-concurrency-limits/limiter"
	"github.com/platinummonkey/go-concurrency-limits/strategy"
	"go.uber.org/zap"
	"tugas-akhir/backend/app/processor/worker"
	"tugas-akhir/backend/infrastructure/amqp"
	entity2 "tugas-akhir/backend/infrastructure/amqp/entity"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/pkg/logger"
)

type Processor struct {
	config        *config.Config
	orderConsumer *amqp.Consumer
	ctx           context.Context
	limiter       core.Limiter
	worker        *worker.BookingWorker
}

func (p *Processor) Run() error {
	// run consume place order
	err := p.ConsumePlaceOrder()

	if err != nil {
		return err
	}

	return nil
}

func (p *Processor) Stop() error {
	l := logger.FromCtx(p.ctx).Sugar()

	if consumeErr := p.orderConsumer.Close(); consumeErr != nil {
		l.Error(consumeErr)
	}

	if publishErr := p.worker.Stop(); publishErr != nil {
		l.Error(publishErr)
	}

	return nil
}

func (p *Processor) ConsumePlaceOrder() error {
	l := logger.FromCtx(p.ctx).With(zap.String("context", "place-order-consumer"))

	go func() {
	connLoop:
		for {
			channel, consumeErr := p.orderConsumer.Consume(p.ctx)

			if consumeErr != nil {
				l.Sugar().Error(consumeErr)
				continue
			}

		mainLoop:
			for {
				select {
				case rawMsg, ok := <-channel:
					if !ok {
						l.Info("updates channel is closed")
						break mainLoop
					}

					l.Info("receiving message")

					go func() {
						listener, ok := p.limiter.Acquire(p.ctx)

						if !ok || listener == nil {
							l.Sugar().Errorf("failed to acquire because not ok or listener is nil")

							if requeueErr := rawMsg.Reject(true); requeueErr != nil {
								l.Sugar().Error(requeueErr)
							}

							if listener != nil {
								listener.OnDropped()
							}

							return
						}

						processErr := p.worker.Process(p.ctx, &rawMsg)

						if processErr != nil {
							listener.OnDropped()
						} else {
							listener.OnSuccess()
						}
					}()
				case <-p.ctx.Done():
					l.Info("consume place order context is done")
					break connLoop
				}
			}
		}
	}()

	return nil
}

func NewProcessor(
	config *config.Config,
	ctx context.Context,
	worker *worker.BookingWorker,
) (*Processor, error) {
	orderConsumer := amqp.NewConsumer(
		config,
		entity.PlaceOrderQueue,
		&entity.PlaceOrderExchange,
		entity2.ConsumeConfig{
			PrefetchCount: 4,
			PrefetchSize:  0,
			AutoAck:       false,
			RoutingKeys:   []string{"orders"},
		},
	)

	// Setup concurrency limits
	// todo integrate with prometheus metrics registry
	// todo update logger to match zap
	limitStrategy := strategy.NewSimpleStrategy(100)

	defaultLimiter, err := limiter2.NewDefaultLimiterWithDefaults(
		"order_processor_limiter",
		limitStrategy,
		limit.BuiltinLimitLogger{},
		core.EmptyMetricRegistryInstance,
	)

	if err != nil {
		return nil, err
	}

	limiter := limiter2.NewQueueBlockingLimiterFromConfig(defaultLimiter, limiter2.QueueLimiterConfig{})

	return &Processor{
		ctx:           ctx,
		orderConsumer: orderConsumer,
		config:        config,
		limiter:       limiter,
		worker:        worker,
	}, nil
}
