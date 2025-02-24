package utility

import "sync"

type LimitedGroup struct {
	wg    sync.WaitGroup
	ch    chan struct{}
	limit int
}

func NewLimitedGroup(limit int) *LimitedGroup {
	return &LimitedGroup{
		ch:    make(chan struct{}, limit),
		limit: limit,
	}
}

func (lg *LimitedGroup) Add(f func()) {
	lg.wg.Add(1)

	go func() {
		defer lg.wg.Done()
		lg.ch <- struct{}{} // acquire token. this will be blocking if the channel is full
		f()
		<-lg.ch // reset token
	}()
}

func (lg *LimitedGroup) Wait() {
	lg.wg.Wait()
	close(lg.ch)
}
