package early_dropper

import (
	"sync"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/pkg/logger"
)

type lockHolder struct {
	Locks sync.Map
}

var LockHolder = lockHolder{
	Locks: sync.Map{},
}

type LockReleaser struct {
	requestID string
	isRun     bool
	onSuccess func() error
	onFailed  func() error
}

func NewLockReleaser(
	requestID string,
	onSuccess func() error,
	onFailed func() error,
) *LockReleaser {
	lock := LockReleaser{
		requestID: requestID,
		isRun:     false,
		onFailed:  onFailed,
		onSuccess: onSuccess,
	}

	if requestID == "" {
		logger.GetInfo().Warn("Lock releaser registered with empty request id. skipping timeout fallback.")
	} else {
		LockHolder.Locks.Store(requestID, &lock)
	}

	return &lock
}

func (r *LockReleaser) releaseFromHolder() {
	if r.requestID != "" {
		LockHolder.Locks.Delete(r.requestID)
	}
}

func (r *LockReleaser) OnSuccess() error {
	defer r.releaseFromHolder()

	if !r.isRun {
		r.isRun = true
		return r.onSuccess()
	}

	return entity.LockAlreadyReleased
}

func (r *LockReleaser) OnFailed() error {
	defer r.releaseFromHolder()

	if !r.isRun {
		r.isRun = true
		return r.onFailed()
	}

	return entity.LockAlreadyReleased
}
