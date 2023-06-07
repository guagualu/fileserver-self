package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	pool      *redis.Pool
	redisHost = "127.0.0.1:6379"
	redisPass = "root1234"
)

const (
	lockCommand = `if redis.call("GET", KEYS[1]) == ARGV[1] then
    redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
    return "OK"
else
    return redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
end`
	delCommand = `if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end`
)

type Rdb struct {
	RedisPool  *redis.Pool
	ReidsMutex string
}

type WatchRes struct {
	Res bool
	Err error
}

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

func (r Rdb) SetMutex(uuid string, ctx context.Context) (bool, error) {
	//key count 是要输入的参数中key的数量
	lua := redis.NewScript(1, lockCommand)
	conn, err := r.RedisPool.Dial()
	if err != nil {

		return false, err
	}
	// uuid 以及 超时时间
	res, err := redis.String(lua.Do(conn, r.ReidsMutex, uuid, 1500))
	if err != nil {
		return false, err
	}
	if res == "OK" {
		watchRes := make(chan WatchRes)
		go r.watchAndPX(watchRes, ctx, conn, uuid)
	}
	return true, nil

}

func (r Rdb) watchAndPX(watchRes chan WatchRes, ctx context.Context, conn redis.Conn, uuid string) {
	//使用定时器 进行续约
	setExTimer := time.NewTimer(1000)
	for {
		select {
		case <-setExTimer.C:
			redis.Int(conn.Do("SET", r.ReidsMutex, uuid, "PX", "1000"))
		}
	}
}

func (r Rdb) DeleteMutex(uuid string) (bool, error) {
	//key count 是要输入的参数中key的数量
	lua := redis.NewScript(1, delCommand)
	conn, err := r.RedisPool.Dial()
	if err != nil {
		return false, err
	}
	// uuid 以及 超时时间
	res, err := redis.Int(lua.Do(conn, r.ReidsMutex, uuid))
	if err != nil {
		return false, err
	}
	if res == 1 {
		return true, nil
	}
	return false, nil

}
