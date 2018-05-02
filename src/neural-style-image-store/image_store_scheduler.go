package ImageStoreService

import (
	"os"
	"path/filepath"
	"strconv"
)

var (
	// MaxWorker define the size of work
	MaxWorker = os.Getenv("MAX_WORKERS")
	// MaxQueue define the size of the cache queue
	MaxQueue = os.Getenv("MAX_QUEUE")
	// MaxResultQueue define the size of the cached result queue
	MaxResultQueue = os.Getenv("MAX_RESULT_QUEUE")
)

// ImageJob define the azure job upload job
type ImageJob struct {
	UploadImage   Image
	ResultChannel chan UploadResult
}

// JobQueue A buffered channel that we can send work requests on.
var JobQueue chan ImageJob

// Stores define group of image stores services
// Azure storage support multiple parallel  store account
var Stores map[string](*AzureImageStore)

// Done closed channel
var Done chan interface{}

func init() {
	queueSize, err := strconv.Atoi(MaxQueue)
	if err != nil {
		queueSize = 2
	}

	workerSize, err := strconv.Atoi(MaxWorker)
	if err != nil {
		workerSize = 2
	}

	Stores = make(map[string]*AzureImageStore)
	JobQueue = make(chan ImageJob, queueSize)
	Done = make(chan interface{})

	storeDispatcher := NewDispatcher(workerSize)
	storeDispatcher.Run()
}

// Worker represents the worker that executes the job
type Worker struct {
	// WorkerPool define the worker load
	WorkerPool chan chan ImageJob
	// JobChannel define the job cache channel
	JobChannel chan ImageJob
	quit       chan bool

	// ImageStore service
	Store *AzureImageStore
}

// NewWorker generate the new worker
func NewWorker(workerPool chan chan ImageJob) Worker {
	storeWorker := Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan ImageJob),
		quit:       make(chan bool),
		Store:      NewAzureImageStore(),
	}

	Stores[storeWorker.Store.StorageAccount] = storeWorker.Store
	return storeWorker
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w Worker) Start() {
	go func() {
		for {
			// register the current worker into the worker queue.
			w.WorkerPool <- w.JobChannel

			select {
			case imgJob := <-w.JobChannel:
				// we have received a work request.
				var fileURL string
				fileURL, err := w.Store.Save(imgJob.UploadImage)
				if err != nil {
					// Todo: log the failed operation
					imgJob.ResultChannel <- UploadResult{
						UserID:      imgJob.UploadImage.UserID,
						Name:        filepath.Base(imgJob.UploadImage.Location),
						Location:    "",
						ImageID:     imgJob.UploadImage.ImageID,
						UploadError: err,
					}
				} else {
					imgJob.ResultChannel <- UploadResult{
						UserID:      imgJob.UploadImage.UserID,
						Name:        filepath.Base(imgJob.UploadImage.Location),
						Location:    fileURL,
						ImageID:     imgJob.UploadImage.ImageID,
						UploadError: nil,
					}
				}

			case <-w.quit:
				// we have received a signal to stop
				close(w.JobChannel)
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
	WorkerPool chan chan ImageJob
	maxWorker  int
	workers    []Worker
}

// NewDispatcher configure the size of Dispatcher
func NewDispatcher(maxWorkerSize int) *Dispatcher {
	pool := make(chan chan ImageJob, maxWorkerSize)
	return &Dispatcher{WorkerPool: pool, maxWorker: maxWorkerSize}
}

// Run generate the dispatcher
func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorker; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
		d.workers = append(d.workers, worker)
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case <-Done:
			// Stop the worker
			for _, w := range d.workers {
				w.Stop()
			}

			close(JobQueue)
			return
		case img := <-JobQueue:
			// a job request has been received
			go func(job ImageJob) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.WorkerPool
				// dispatch the job to the worker job channel
				jobChannel <- job
			}(img)
		}
	}
}
