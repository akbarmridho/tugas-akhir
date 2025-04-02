package processor

import (
	"context"
	"go.uber.org/fx"
	"tugas-akhir/backend/app/processor/worker"
)

var Module = fx.Options(
	fx.Provide(worker.NewResultPublisher),
	fx.Provide(worker.NewBookingWorker),
	fx.Provide(NewProcessor),
	fx.Provide(fx.Annotate(
		NewMetricsServer,
		fx.OnStop(func(ctx context.Context, metricsServer *MetricsServer) error {
			return metricsServer.Stop(ctx)
		}),
	)),
)
