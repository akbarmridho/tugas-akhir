package early_dropper

import "tugas-akhir/backend/internal/orders/entity"

type LockReleaser struct {
	isRun     bool
	onSuccess func() error
	onFailed  func() error
}

func (r *LockReleaser) OnSuccess() error {
	if !r.isRun {
		r.isRun = true
		return r.onSuccess()
	}

	return entity.LockAlreadyReleased
}

func (r *LockReleaser) OnFailed() error {
	if !r.isRun {
		r.isRun = true
		return r.onFailed()
	}

	return entity.LockAlreadyReleased
}
