package util

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"go.uber.org/zap"
	"im_socket_server/constant"
	"im_socket_server/dao"
	"im_socket_server/logs"
	"im_socket_server/mode"
)

/***
获取用户基本信息
*/
func GetUserInfo(userId string) mode.User {
	logs.Loggers.Info("GetUserInfo", zap.String("userId", userId))
	conn := dao.RedisDefaultPool.Get()
	redisJson, err := redis.String(conn.Do("get", constant.USER_INFO+userId))
	if err != nil {
		logs.Loggers.Error("GetUserInfo-redisJson-err", zap.Error(err))
	}
	userInfo := &mode.User{}
	json.Unmarshal([]byte(redisJson), userInfo)
	return *userInfo
}
