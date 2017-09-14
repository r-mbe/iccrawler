package main

import (
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
			fmt.Println(time.Now())
		case <-stop:
			return
		}
	}
}
