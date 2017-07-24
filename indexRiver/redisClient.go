package main

import (
	"fmt"
	"flag"
	"github.com/stanxii/indexRiver/redis"
)


var (
    cacheKey      = "onlineClient"
    redisServer   = flag.String("redisServer", "127.0.0.1:6379", "")
    redisPassword = flag.String("redisPassword", "", "")
    userId        = flag.String("userId", "", "")
    kickCmd       = flag.String("kick", "", "")
    clearCmd      = flag.String("clear", "", "")
)

func main() {

	flag.Parse()

	client := redis.NewClient(*redisServer, *redisPassword)
    defer client.Close()

	supplier, err := client.BRPop("list.ickey", 10)
    if err != nil{
        fmt.Println("rpop got err.stan")
    }

	fmt.Println("rpop resutl=: %s" , supplier)

}


