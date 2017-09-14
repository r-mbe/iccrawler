package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context) error {

	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			fmt.Println("finished.")
			return
		}
	}(ctx)

	for {
		fmt.Println("### Default do work worker## ", time.Now())

		tick := time.NewTicker(3000 * time.Second)
		select {
		case <-tick.C:
			fmt.Println("too long time")
		}
	}

}

func main() {
	stop := time.After(3 * 10 * time.Second)

	dur := 5
	//tick for one day run once worker
	tick := time.NewTicker(time.Duration(dur) * time.Second)
	defer tick.Stop()
	for {

		durS := 2*dur - 1
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(durS)*time.Second)
		select {
		case <-stop:
			fmt.Println("all done.")
			cancel()
			return
		case <-tick.C:
			go worker(ctx)
		}
	}

}
