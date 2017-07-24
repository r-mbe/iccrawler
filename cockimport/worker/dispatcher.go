package worker

type Dispatcher struct {
	MaxWorkers int
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan Job
	Workers    []Worker
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{WorkerPool: pool, MaxWorkers: maxWorkers, Workers: make([]Worker, maxWorkers)}
}

func (d *Dispatcher) Run(queues int) chan Job {
	// starting n number of workers
	for i := 0; i < d.MaxWorkers; i++ {
		d.Workers[i] = NewWorker(d.WorkerPool)
		d.Workers[i].Start()
	}

	JobQueue := make(chan Job, queues)

	go d.dispatch(JobQueue)

	return JobQueue
}

func (d *Dispatcher) Stop() {
	for i := 0; i < d.MaxWorkers; i++ {
		d.Workers[i].Stop()
	}
}

func (d *Dispatcher) dispatch(jQ chan Job) {
	for {
		select {
		case job := <-jQ:
			// a job request has been received
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.WorkerPool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		}
	}
}
