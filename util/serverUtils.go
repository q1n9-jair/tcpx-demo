package util

import (
	"github.com/garyburd/redigo/redis"
	"go.uber.org/zap"
	"im_socket_server/config"
	"im_socket_server/constant"
	"im_socket_server/dao"
	"im_socket_server/logs"
	"time"
)

var ServerNameKey string

//服务注册
func ServerRegistryCenter() {
	c := config.GetConfig()
	host := c.GetString("registryCenter.host")
	port := c.GetString("registryCenter.port")
	name := c.GetString("registryCenter.name")
	appCode := c.GetString("registryCenter.appCode")
	channel := c.GetString("registryCenter.channel")
	domain := c.GetString("registryCenter.domain")
	registryCenterMap := make(map[string]interface{})
	registryCenterMap["host"] = host
	registryCenterMap["port"] = port
	registryCenterMap["name"] = name
	registryCenterMap["appCode"] = appCode
	registryCenterMap["channel"] = channel
	registryCenterMap["domain"] = domain
	registryCenterMap["livesTime"] = time.Now().Unix() //设置续命时间
	ServerNameKey = constant.SERVER_REGISTRY_INFO_KEY + name
	dao.SetMap(ServerNameKey, registryCenterMap)
	//放到在线服务器集合
	conn := dao.RedisDefaultPool.Get()
	_, serverRegistryCenterErr := conn.Do("sadd", constant.SERVER_REGISTRY_CENTER_KEY, ServerNameKey)
	if serverRegistryCenterErr != nil {
		logs.Loggers.Error("serverRegistryCenterErr:", zap.Error(serverRegistryCenterErr))
	}
}

//服务续命
func ServerLives() {
	conn := dao.RedisDefaultPool.Get()
	//从服务器注册集合检查这个值
	serverIsmembers, ServerRegistryCenterErr := redis.Int(conn.Do("SISMEMBER", constant.SERVER_REGISTRY_CENTER_KEY, ServerNameKey))
	if ServerRegistryCenterErr != nil {
		logs.Loggers.Error("serverRegistryCenterErr:", zap.Error(ServerRegistryCenterErr))
	}
	serverHexists, serverHexistsErr := redis.Int(conn.Do("HEXISTS", ServerNameKey, "name"))
	if serverHexistsErr != nil {
		logs.Loggers.Error("serverHexistsErr:", zap.Error(serverHexistsErr))
	}
	//判断这个key存不存在 ，不存在就重新注册，存在就续命
	if serverIsmembers == 0 && serverHexists == 0 {
		ServerRegistryCenter()
	} else {
		_, err := conn.Do("hset", ServerNameKey, "livesTime", time.Now().Unix())
		if err != nil {
			logs.Loggers.Error("ServerLives-ServerIsmembers-err:", zap.Error(err))
		}
	}

}
