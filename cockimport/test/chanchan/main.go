package main

import "fmt"

var (
	g    = make(chan int)
	quit = make(chan chan bool)
)

func main() {

	go B()
	for i := 0; i < 5; i++ {
		g <- i
	}
	wait := make(chan bool)
	quit <- wait
	<-wait //这样就可以等待B的退出了
	fmt.Println("Main Quit")
}

func B() {
	for {
		select {
		case i := <-g:
			fmt.Println(i + 1)
		case c := <-quit:
			c <- true
			fmt.Println("B Quit")
			return
		}
	}
}
