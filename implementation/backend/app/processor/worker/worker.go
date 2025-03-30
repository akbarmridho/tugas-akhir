package worker

type BookingWorker struct {
	// todo
}

// Process will synchronously perform a job and return the result.
func (w *BookingWorker) Process(rawPayload interface{}) interface{} {
	//payload := rawPayload.(entity.PlaceOrderDto)
	return nil
}

// BlockUntilReady is called before each job is processed and must block the
// calling goroutine until the Worker is ready to process the next job.
func (w *BookingWorker) BlockUntilReady() {
	// todo
}

// Interrupt is called when a job is cancelled. The worker is responsible
// for unblocking the Process implementation.
func (w *BookingWorker) Interrupt() {
	// todo
}

// Terminate is called when a Worker is removed from the processing pool
// and is responsible for cleaning up any held resources.
func (w *BookingWorker) Terminate() {
	// todo
}

func NewBookingWorker() *BookingWorker {
	return &BookingWorker{}
}
