package ImageStoreService

import (
	"os"
	"strconv"
)

var (
	// MaxWorker define the size of work
	MaxWorker = os.Getenv("MAX_WORKERS")
	// MaxQueue define the size of the cache queue
	MaxQueue = os.Getenv("MAX_QUEUE")
)

// JobQueue A buffered channel that we can send work requests on.
var JobQueue chan ImageStore

func init() {
	queueSize, _ := strconv.Atoi(MaxQueue)
	workerSize, _ := strconv.Atoi(MaxWorker)

	JobQueue = make(chan ImageStore, queueSize)

	storeDispatcher := NewDispatcher(workerSize)
	storeDispatcher.Run()
	storeDispatcher.dispatch()
}

// Worker represents the worker that executes the job
type Worker struct {
	// WorkerPool define the worker load
	WorkerPool chan chan ImageStore
	// JobChannel define the job cache channel
	JobChannel chan ImageStore
	quit       chan bool
}

// NewWorker generate the new worker
func NewWorker(workerPool chan chan ImageStore) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan ImageStore),
		quit:       make(chan bool)}
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w Worker) Start() {
	go func() {
		for {
			// register the current worker into the worker queue.
			w.WorkerPool <- w.JobChannel

			select {
			case imgStore := <-w.JobChannel:
				// we have received a work request.
				if err := imgStore.Save(nil); err != nil {

				}

			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

// Dispatcher job schedule
type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan ImageStore
	maxWorker  int
}

// NewDispatcher configure the size of Dispatcher
func NewDispatcher(maxWorkerSize int) *Dispatcher {
	pool := make(chan chan ImageStore, maxWorkerSize)
	return &Dispatcher{WorkerPool: pool, maxWorker: maxWorkerSize}
}

// Run generate the dispatcher
func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorker; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-JobQueue:
			// a job request has been received
			go func(job ImageStore) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.WorkerPool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		}
	}
}
