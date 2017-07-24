package main

import (
	"fmt"
	"strconv"
)

func main() {
	var chs1 = make(chan int)
	var chs2 = make(chan float64)
	var chs3 = make(chan string)
	var ch4close = make(chan int)
	defer close(ch4close)

	go func(c chan int, ch4close chan int) {
		for i := 0; i < 5; i++ {
			c <- i
		}
		close(c)
		ch4close <- 1
	}(chs1, ch4close)

	go func(c chan float64, ch4close chan int) {
		for i := 0; i < 5; i++ {
			c <- float64(i) + 0.1
		}
		close(c)
		ch4close <- 1
	}(chs2, ch4close)

	go func(c chan string, ch4close chan int) {
		for i := 0; i < 5; i++ {
			c <- "string:" + strconv.Itoa(i)
		}
		close(c)
		ch4close <- 1
	}(chs3, ch4close)

	done := 0
	finished := 0
	for finished < 3 {
		select {
		case v, ok := <-chs1:
			if ok {
				done = done + 1
				fmt.Println(0, v, "chan....1")
			}
		case v, ok := <-chs2:
			if ok {
				done = done + 1
				fmt.Println(1, v, "chan....2.")
			}
		case v, ok := <-chs3:
			if ok {
				done = done + 1
				fmt.Println(2, v, "chan....3.")
			}
		case _, ok := <-ch4close:
			if ok {
				finished = finished + 1
			}
		}
	}
	fmt.Println("Done", done)
}
