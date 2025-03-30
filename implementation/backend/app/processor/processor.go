package processor

import (
	"context"
	"github.com/Jeffail/tunny"
	"go.uber.org/zap"
	"runtime"
	"tugas-akhir/backend/app/processor/worker"
	"tugas-akhir/backend/infrastructure/amqp"
	entity2 "tugas-akhir/backend/infrastructure/amqp/entity"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/orders/usecase/place_order"
	"tugas-akhir/backend/pkg/logger"
)

type Processor struct {
	pool            *tunny.Pool
	config          *config.Config
	orderConsumer   *amqp.Consumer
	resultPublisher *worker.ResultPublisher
	ctx             context.Context
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

	p.pool.Close()

	if publishErr := p.resultPublisher.Stop(); publishErr != nil {
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
						_, processErr := p.pool.ProcessCtx(p.ctx, rawMsg)

						if processErr != nil {
							l.Sugar().Error(processErr)
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
	placeOrderUsecase place_order.PlaceOrderUsecase,
) *Processor {
	numCPU := runtime.NumCPU()

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

	resultPublisher := worker.NewResultPublisher(config)

	pool := tunny.New(numCPU, func() tunny.Worker {
		return worker.NewBookingWorker(ctx, placeOrderUsecase, resultPublisher)
	})

	// todo ability to adjust flow rate

	return &Processor{
		ctx:             ctx,
		pool:            pool,
		orderConsumer:   orderConsumer,
		resultPublisher: resultPublisher,
		config:          config,
	}
}
