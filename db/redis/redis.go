package redis

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	pool      *redis.Pool
	redisHost = "127.0.0.1:6379"
	redisPass = "root1234"
)

//初始化启动 单例模式
func init() {
	pool = newRedisPool()
}

func RedisPool() *redis.Pool {
	return pool
}

//创建redis连接池
func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,                //最大连接数
		MaxActive:   0,                 //最多可用连接<maxidle
		IdleTimeout: 300 * time.Second, //连接超过多久没用就算超时
		Dial: func() (redis.Conn, error) {
			//1、打开连接
			conn, err := redis.Dial("tcp", redisHost)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			//2、访问认证
			if _, err = conn.Do("AUTH", redisPass); err != nil {
				conn.Close()
				return nil, err
			}
			return conn, nil

		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute { //?????????
				return nil
			}
			_, err := c.Do(
				"PING",
			)
			if err != nil {
				return err
			}
			return nil
		}, //检查reidsserver的健康状况，不健康就断开连接
	}
}
