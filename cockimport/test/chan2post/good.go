// Command pubsub is an example of a fanout exchange with dynamic reliable
// membership, reading from stdin, writing to stdout.
//
// This example shows how to implement reconnect logic independent from a
// publish/subscribe loop with bridges to application types.

package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"

	"context"
)

var (
	MaxWorker = 5000
	MaxQueue  = 100000
)

// Lighting main struct
// type Lighting struct {
// 	JobQueue chan Payload
// }

type Payload struct {
	I    int
	Name string
}

func (p *Payload) WorkingHard(ctx context.Context) error {
	fmt.Println("faking.... real storage now.")

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	select {
	case <-ctx.Done():
		return errors.New("parent cancel all.")
	case <-cctx.Done():
		fmt.Println("storage don. :w")
		return nil
	default:
		//DoWorkingHard()
		fmt.Println("DoWorkingHard doing.......")

	}

	return nil
}

// Job represents the job to be run
type Job struct {
	Payload Payload
}

// JobQueue A buffered channel that we can send work requests on.
var JobQueue chan Job

// worker represents the worker that executes the job
type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	ctx        context.Context
	quit       chan bool
}

func NewWorker(ctx context.Context, workerPool chan chan Job) *Worker {
	return &Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		ctx:        ctx,
		quit:       make(chan bool),
	}
}

//
// func NewLighting() * Lighting {
// 	return &Lighting{
// 		JobQueue:
// 	}
// }
// func init() {
//
// }

func (w *Worker) Start() {
	go func(w *Worker) {

		for {
			// register the current worker into the worker queue.
			w.WorkerPool <- w.JobChannel

			select {
			case <-w.ctx.Done():
				fmt.Println("Parent w.ctx.Done working hard all. done.")
				return
			case job := <-w.JobChannel:
				//?? confuse why not <- w.WorkerPool
				// we hae received a worker request.
				if err := job.Payload.WorkingHard(w.ctx); err != nil {
					// log.Print("Error working hard %s\n", err.Error())
					fmt.Printf("Error working hard %s\n", err.Error())
				}

			case <-w.quit:
				fmt.Println("receive a singal quit to stop don. :w")
				return
			}
		}
	}(w)

}

func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func PayloadHandler(jq chan<- Job) {

	var payloads [3]Payload
	// payloads := []Payload{{3, "stan"}, {4, "xiyx"}, {5, "sss"}, {4, "xiyx"}, {5, "sss"}}
	for i := 0; i < 2; i++ {
		payloads[i].I = i
		payloads[i].Name = "aa" + strconv.Itoa(i)
	}

	for _, payload := range payloads {
		// let's creat a job with the payload
		job := Job{Payload: payload}

		// Push the work onto the queue.
		fmt.Println("send job to job queue.")
		jq <- job
	}
}

type Dispatcher struct {
	// A pool of workers channels that are registered with the Dispatcher
	WorkerPool chan chan Job
	MaxWorkers int
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	ctx, cancel := context.WithCancel(context.Background())

	return &Dispatcher{WorkerPool: pool,
		MaxWorkers: maxWorkers,
		ctx:        ctx,
		cancel:     cancel,
	}

}

func (d *Dispatcher) Run() {
	// starting n number of workers.
	for i := 0; i < d.MaxWorkers; i++ {
		worker := NewWorker(d.ctx, d.WorkerPool)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) WaitEnd() {
	select {
	case <-d.ctx.Done():
		fmt.Println("All Done.")
		return
	}
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-JobQueue:
			// a job request has been received
			fmt.Println(" a job request has been received")
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				//this will block until a worker is idle
				fmt.Println("before jobChannel := <-d.WorkerPool this will block until a worker is idle")
				jobChannel := <-d.WorkerPool
				fmt.Println("end jobChannel := <-d.WorkerPool this will block until a worker is idle")

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		}
	}
}

func (d *Dispatcher) Stop() {
	d.cancel()
}

func main() {
	fmt.Println("hello context chan chan!")

	var worker = flag.Int("worker", 5000, "max worker for queue pool")
	var queue = flag.Int("queue", 100000, "max queue number chanchan")

	flag.Parse()

	fmt.Println("worker:", *worker)
	fmt.Println("queue:", *queue)

	JobQueue = make(chan Job, *queue)
	defer close(JobQueue)

	d := NewDispatcher(*worker)

	d.Run()

	PayloadHandler(JobQueue)

	for i := 0; i < 50000; i++ {
		PayloadHandler(JobQueue)
	}
	//
	// for i := 0; i < *worker; i++ {
	// 	<-d.done
	// }

	// select {
	// case <-time.After(time.Second * 3):
	// 	d.Stop()
	// }

	defer d.Stop()

	go d.WaitEnd()

	fmt.Println("Main Quit")

}
