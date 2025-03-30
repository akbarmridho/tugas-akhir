package processor

import (
	"go.uber.org/fx"
	"tugas-akhir/backend/app/processor/worker"
)

var Module = fx.Options(
	fx.Provide(worker.NewResultPublisher),
	fx.Provide(NewProcessor),
)
