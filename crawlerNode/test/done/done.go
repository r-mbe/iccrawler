package main

import (
	"fmt"
	"time"
)

// In this example we'll use a `jobs` channel to
// communicate work to be done from the `main()` goroutine
// to a worker goroutine. When we have no more jobs for
// the worker we'll `close` the `jobs` channel.
func main() {
	jobs := make(chan int)
	pages := make(chan int)
	done := make(chan struct{})

	// Here's the worker goroutine. It repeatedly receives
	// from `jobs` with `j, more := <-jobs`. In this
	// special 2-value form of receive, the `more` value
	// will be `false` if `jobs` has been `close`d and all
	// values in the channel have already been received.
	// We use this to notify on `done` when we've worked
	// all our jobs.
	start := time.Now()
	//pages
	go func() {
		defer close(done)
		for {
			j, more := <-pages
			if more {
				fmt.Println("received pages", j)
			} else {
				fmt.Println("received all pages")
				//done <- true
				return
			}
		}
	}()

	//jobs
	go func() {
		defer close(pages)
		for {
			j, more := <-jobs
			if more {
				fmt.Println("received job", j)
				pages <- j
			} else {
				fmt.Println("received all jobs")
				//done <- true
				return
			}
		}
	}()

	// This sends 3 jobs to the worker over the `jobs`
	// channel, then closes it.
	// var wg sync.WaitGroup
	// for j := 1; j <= 50000; j++ {
	// 	go func(j int) {
	// 		wg.Add(1)
	// 		defer wg.Done()
	// 		jobs <- j
	// 		fmt.Println("sent job", j)
	// 	}(j)
	// }
	// wg.Wait()

	////////////////////////// 1w goroutine
	// channel, then closes it.
	for j := 1; j <= 300000; j++ {
		go func(j int) {
			jobs <- j
			fmt.Println("sent job", j)
		}(j)
	}

	// var wg sync.WaitGroup
	// go func() {
	// 	var wg sync.WaitGroup
	// 	for j := 1; j <= 300000; j++ {
	// 		go func(j int) {
	// 			wg.Add(1)
	// 			defer wg.Done()
	// 			jobs <- j
	// 			fmt.Println("sent job", j)
	// 		}(j)
	// 	}
	// 	wg.Wait()
	// 	close(jobs)
	// }()

	/////////////////////////////////
	// go func() {
	// 	for j := 1; j <= 300000; j++ {
	// 		jobs <- j
	// 		fmt.Println("sent job", j)
	// 	}
	// 	defer close(jobs)
	// }()

	for j := 1; j <= 300000; j++ {
		jobs <- j
		fmt.Println("sent job", j)
	}
	close(jobs)
	<-done

	dur := time.Since(start).Seconds()
	fmt.Printf("sent all jobs. spendtime: %v\n", dur)

	// We await the worker using the
	// [synchronization](channel-synchronization) approach
	// we saw earlier.

}
