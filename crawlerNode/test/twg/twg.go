package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	urls := []string{
		"http://www.golang.com",
		"http://www.google.com",
		"http://www.baiyuxiong.com",
		"http://www.baiyuxiong.com",
		"http://www.baiyuxiong.com",
		"http://www.baiyuxiong.com",
		"http://www.baiyuxiong.com",
		"http://www.baiyuxiong.com",
	}

	for _, url := range urls {
		go func(url string) {
			wg.Add(1)
			defer wg.Done()

			res, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(res.StatusCode)
		}(url)
	}

	wg.Wait()
	fmt.Println("game over.")
}
