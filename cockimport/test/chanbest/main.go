package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"techtoolkit.ickey.cn/cockimport/job"
	"techtoolkit.ickey.cn/cockimport/worker"
)

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var jobs = flag.Int("j", 100, "Number of jobs")
	var workers = flag.Int("w", runtime.NumCPU(), "Number of workers")
	var queues = flag.Int("q", 2, "Number of queues")
	var fail = flag.Bool("f", false, "Fail randomly")
	var retries = flag.Int("r", 1, "Number of retries for failed jobs")
	flag.Parse()

	dispatcher := worker.NewDispatcher(*workers)
	JobQueue := dispatcher.Run(*queues)

	wg := &sync.WaitGroup{}
	for i := 0; i < *jobs; i++ {
		wg.Add(1)
		fmt.Printf("Adding job %d to the queue\n", i)
		JobQueue <- job.Job{Name: fmt.Sprintf("%d", i), Wg: wg, Rnd: r, JobQueue: JobQueue, RandomFail: *fail, Retries: *retries}
	}

	wg.Wait()
	dispatcher.Stop()
}
