package dao

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"strconv"
	"tcpx-demo/config"
	"time"
)

var RedisDefaultPool *redis.Pool

func newPool(addr, pwd, db string) *redis.Pool {
	dbNum, err := strconv.Atoi(db)
	if err != nil {
		log.Print(err)
	}
	return &redis.Pool{
		MaxIdle:     300,
		MaxActive:   0,
		IdleTimeout: 240 * time.Second,
		Dial: func() (conn redis.Conn, e error) {
			return redis.Dial("tcp", addr, redis.DialDatabase(dbNum), redis.DialPassword(pwd))
		},
	}
}

func init() {
	c := config.GetConfig()
	redisHost := c.GetString("redis.host")
	redisPwd := c.GetString("redis.pwd")
	redisDB := c.GetString("redis.RedisDB")
	RedisDefaultPool = newPool(redisHost, redisPwd, redisDB)
}
