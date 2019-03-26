package g

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

var RedisConnPool *redis.Pool

// 初始化redis连接池
func InitRedisConnPool() {
	redisConfig := Config().Redis

	RedisConnPool = &redis.Pool{
		MaxIdle:     redisConfig.MaxIdle,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisConfig.Addr)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: PingRedis,
	}
}

func PingRedis(c redis.Conn, t time.Time) error {
	_, err := c.Do("ping")
	if err != nil {
		log.Println("[ERROR] ping redis fail", err)
	}
	return err
}
