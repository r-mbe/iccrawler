package redis

import (
	//	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

/*
var (
    cacheKey      = "onlineClient"
    redisServer   = flag.String("redisServer", "172.18.0.2:6379", "")
    redisPassword = flag.String("redisPassword", "", "")
    userId        = flag.String("userId", "", "")
    kickCmd       = flag.String("kick", "", "")
    clearCmd      = flag.String("clear", "", "")
)
*/

type RedisClient struct {
	pool *redis.Pool
}

//获取所有在线用户
func (c *RedisClient) GetAll(key string) {
	conn := c.pool.Get()
	defer conn.Close()

	clients, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		panic(err)
	}
	fmt.Printf("online client: %d \n", len(clients))
	for uId, client := range clients {
		fmt.Printf("%s -- %s\n", uId, client)
	}
}

//根据用户ID获取单个用户
func (c *RedisClient) GetOne(key, id string) {
	conn := c.pool.Get()
	defer conn.Close()

	client, err := redis.String(conn.Do("HGET", key, id))

	if err != nil {
		panic(err)
	}
	fmt.Println(client)
}

//踢出某个用户
func (c *RedisClient) Kick(key string, id string) {
	conn := c.pool.Get()
	defer conn.Close()

	result, err := c.pool.Get().Do("HDEL", key, id)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

//清除所有在线用户信息
func (c *RedisClient) ClearAll() {
	conn := c.pool.Get()
	defer conn.Close()

	key := "onlineClient"

	result, err := conn.Do("DEL", key)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

//////////stan begin

// Block and Right Pop an element off the queue
// Use timeout_secs = 0 to block indefinitely
// On timeout, this DOES return an error because redigo does.
// brpop response is a []byte   result is arr[1]
//172.18.0.2:6379> brpop list.ickey 10
//1) "list.ickey"
//2) "aii"

func (c *RedisClient) BRPop(key string, timeout_secs int) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()

	var data string

	reply, err := redis.Values(conn.Do("BRPOP", key, timeout_secs))
	// BRPOP会在timeout秒内取出数据，否则返回错误提示
	if err != nil {
		//logp.Info("redis Brpop get err====%v", err)
		if -1 != strings.LastIndexAny(err.Error(), "nil returned") {
			//判断err返回的消息内容，redigo中BRPOP无内容返回会返回"redigo: nil returned"
			//可以模拟为timeout
			return "timeout", errors.New("timeout")
		} else {
			return "err", err
		}
	}
	// 没有错误返回则说明有数据返回，将数据解析传回需要的地方
	redis.Scan(reply, &data, &data)
	// parse data, return if necessary
	//fmt.Println("got final string data: " + data)

	/*
	      another
	   	if res != nil && err == nil {
	   		if arr, ok := res.([]interface{}); ok && len(arr) == 2 {
	   			if bres, ok := arr[1].([]byte); ok {
	   				return bres, err
	   			}
	   		}
	   	}
	*/

	return data, nil
}

/////////////////stan end

//关闭redis连接池
func (c *RedisClient) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

func (c *RedisClient) ActiveCount() int {
	if c.pool != nil {
		return c.pool.ActiveCount()
	} else {
		return 0
	}
}

func NewClient(server, password string) *RedisClient {
	return &RedisClient{
		pool: newPool(server, password),
	}
}

//创建redis connection pool
func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}

			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
