package main

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

func main() {
	go NewSubscriber("wishing")
	go NewSubscriber("downloading")
	go NewSubscriber("crawlering")
	NewSubscriber("indexing")
}

func NewSubscriber(channel string) {
	for {
		fmt.Println("Go redis subscribe Listening.... channel: " + channel)
		c, err := redis.Dial("tcp", "10.8.15.191:6379")
		if err != nil {
			fmt.Println("error connection to redis")
			time.Sleep(5 * time.Second)
			continue
		}
		psc := redis.PubSubConn{c}
		psc.Subscribe(channel)
	ReceiveLoop:
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
			case redis.Subscription:
				fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
			case error:
				fmt.Println("there was an error")
				fmt.Println(v)
				time.Sleep(5 * time.Second)
				psc.Close()
				break ReceiveLoop
			}
		}
	}
}
