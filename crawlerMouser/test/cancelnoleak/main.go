package main

import (
	"context"
	"fmt"
	"time"
)

func main() {

	stop := time.After(3 * 10 * time.Second)
	tick := time.NewTicker(3 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			fmt.Println("Looping.....................", time.Now())
			fmt.Println("Program started...")
			start := time.Now()

			// Initialise top-level context
			ctx := context.Background()

			operationWithTimeout(ctx)

			elapsed := time.Since(start)
			fmt.Printf("Program finished... It took %s \n", elapsed)

		case <-stop:
			fmt.Println("#################All Loop done")
			return
		}
	}

}

func operationWithTimeout(ctx context.Context) {
	fmt.Println("operationWithTimeout started...")

	// Create a channel for signal handling
	c := make(chan bool)
	defer close(c)

	// Define a cancellation after 1s in the context
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Run slowOperation via a goroutine
	go func() {
		slowOperation(c)
	}()

	// Listening to signals
	select {
	case <-ctx.Done():
		fmt.Println(ctx.Err())

	case <-c:
		fmt.Println("Unexpected success!")
	}

	fmt.Println("operationWithTimeout finished...")
}

func slowOperation(c chan bool) {
	fmt.Println("slowOperation started...")

	time.Sleep(4 * time.Second)

	c <- true

	fmt.Println("slowOperation finished...")
}

func gen(ctx context.Context) <-chan int {
	dst := make(chan int)
}
