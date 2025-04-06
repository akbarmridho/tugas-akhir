package processor

import (
	"context"
	"github.com/platinummonkey/go-concurrency-limits/core"
	"go.uber.org/zap"
	"time"
	"tugas-akhir/backend/app/processor/worker"
	"tugas-akhir/backend/infrastructure/amqp"
	entity2 "tugas-akhir/backend/infrastructure/amqp/entity"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/internal/orders/entity"
	go_metrics_prometheus "tugas-akhir/backend/pkg/go-metrics-prometheus"
	"tugas-akhir/backend/pkg/logger"
)

const PollInterval = 3 * time.Second
const ProcessorNamespace = string(config.FlowControlVariant__DropperAsync)
const ProcessorSubsystem = "order_processor"
const LimiterName = "order_processor_limiter"
const StrategyLimit = 1000
const ConcurrencyLimit = 1000

type Processor struct {
	config           *config.Config
	orderConsumer    *amqp.Consumer
	ctx              context.Context
	limiter          core.Limiter
	worker           *worker.BookingWorker
	prometheusClient *go_metrics_prometheus.PrometheusConfig
}

func (p *Processor) Run() error {
	go p.prometheusClient.UpdatePrometheusMetrics()

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
						l.Info("channel is closed")
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
			PrefetchCount: ConcurrencyLimit,
			PrefetchSize:  0,
			AutoAck:       false,
			RoutingKeys:   []string{"orders"},
		},
	)

	limiter, prometheusClient, err := NewLimiter(ctx)

	if err != nil {
		return nil, err
	}

	return &Processor{
		ctx:              ctx,
		orderConsumer:    orderConsumer,
		config:           config,
		limiter:          limiter,
		worker:           worker,
		prometheusClient: prometheusClient,
	}, nil
}
