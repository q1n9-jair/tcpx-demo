package dao

import (
	"github.com/garyburd/redigo/redis"
	"go.uber.org/zap"
	"im_socket_server/config"
	"im_socket_server/logs"
	"log"
	"strconv"
	"time"
)

var RedisDefaultPool *redis.Pool

func newPool(addr, pwd, db string) *redis.Pool {
	dbNum, err := strconv.Atoi(db)
	if err != nil {
		log.Print(err)
	}
	return &redis.Pool{
		MaxIdle:     3,
		MaxActive:   0,
		IdleTimeout: 240 * time.Second,
		Dial: func() (conn redis.Conn, e error) {
			return redis.Dial("tcp", addr, redis.DialDatabase(dbNum), redis.DialPassword(pwd))
		},
	}
}

func SetMap(key string, val map[string]interface{}) bool {
	conn := RedisDefaultPool.Get()
	for k, v := range val {
		_, setMapErr := conn.Do("hset", key, k, v)
		if setMapErr != nil {
			logs.Loggers.Error("SetMap:", zap.Error(setMapErr))
			return false
		}
	}
	return true
}

func init() {
	c := config.GetConfig()
	redisHost := c.GetString("redis.host")
	redisPwd := c.GetString("redis.pwd")
	redisDB := c.GetString("redis.RedisDB")
	RedisDefaultPool = newPool(redisHost, redisPwd, redisDB)
}
