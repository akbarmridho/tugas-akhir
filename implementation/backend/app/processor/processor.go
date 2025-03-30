package processor

import (
	"context"
	"github.com/Jeffail/tunny"
	"runtime"
	"tugas-akhir/backend/app/processor/worker"
)

type Processor struct {
	Pool *tunny.Pool
	Ctx  context.Context
}

func (p *Processor) Run(ctx context.Context) error {
	p.Ctx = ctx

	// todo

	return nil
}

func (p *Processor) Stop(ctx context.Context) error {
	// todo
	// 	s.Scheduler.Stop()
	//	s.Scheduler.Wait(ctx)
	p.Pool.Close()
	return nil
}

func NewProcessor() *Processor {
	numCPU := runtime.NumCPU()

	pool := tunny.New(numCPU, func() tunny.Worker {
		// todo pass required dependency here
		return worker.NewBookingWorker()
	})

	return &Processor{
		Pool: pool,
	}
}
