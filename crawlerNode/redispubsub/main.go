package main

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

func main() {
	for {
		fmt.Println("connecting...")
		c, err := redis.Dial("tcp", "localhost:6379")
		if err != nil {
			fmt.Println("error connecting to redis")
			time.Sleep(5 * time.Second)
			continue
		}
		psc := redis.PubSubConn{c}
		psc.Subscribe("wishchan")
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
