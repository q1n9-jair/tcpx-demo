package util

import (
	"github.com/garyburd/redigo/redis"
	"go.uber.org/zap"
	"im_socket_server/constant"
	"im_socket_server/dao"
	"im_socket_server/logs"
	"im_socket_server/mode"
	"strconv"
	"sync"
	"time"
)

var userStatusWg sync.WaitGroup

func SetOfflineUser(userId string) {
	defer func() {
		errs := recover()
		if errs != nil {
			logs.Loggers.Error("SetOfflineUser:", zap.Reflect("err", errs))
		}
	}()
	//设置redis
	conn := dao.RedisDefaultPool.Get()
	//获取这个人是否为机器人
	isBotUser, botErr := redis.Bool(conn.Do("SISMEMBER", constant.BOT_USER_LIST, userId))
	if isBotUser {
		return
	}
	if botErr != nil {
		logs.Loggers.Error("botErr", zap.Error(botErr))
	}
	//设置最后登陆时间错
	now := time.Now()
	nowUnix := strconv.FormatInt(now.UnixNano()/1e6, 10)

	//获取用户的隐私设置
	var redisKey = constant.USER_PRIVACY_SET_KEY + userId
	ret, _ := redis.String(conn.Do("get", redisKey))

	//如果是不存在隐私设置的 或者是公开的 才修改在线状态
	userPrivacySet := &mode.UserPrivacySet{}
	if ret != "" {
		userPrivacySet = mode.GetUserPrivacySet(ret)
	} else {
		userPrivacySet = nil
	}

	//设置es下线用户
	isUpdateEsOffUser := UpdateEsOffUser(userId, nowUnix)
	logs.Loggers.Info("SetOfflineUser-修改es状态", zap.Bool("isUpdateEsOffUser", isUpdateEsOffUser))
	if !isUpdateEsOffUser {
		logs.Loggers.Info("SetOfflineUser-修改es状态失败", zap.Bool("isUpdateEsOffUser", isUpdateEsOffUser))
		return
	}
	//如果为空或者等于1 就可以设置
	if userPrivacySet == nil || userPrivacySet.OnlineStatus == 1 {
		DelUserOnlineServer(userId)
		//在redis上设置最后登录时间
		_, redisSetLoginLastErr := conn.Do("HSET", constant.USER_LAST_LOGIN_DATE, userId, nowUnix)
		if redisSetLoginLastErr != nil {
			logs.Loggers.Error("redisSetLoginLastErr:", zap.Error(redisSetLoginLastErr))
		}
		_, redisOlDelErr := conn.Do("HDEL", constant.USER_ONLINE_LIST, userId)
		if redisOlDelErr != nil {
			logs.Loggers.Error("redisOlDelErr:", zap.Error(redisOlDelErr))
		}
		logs.Loggers.Info("user.online.list", zap.String("hdel", "userId:"+userId))
	}
	userStatusWg.Done()
}

//设置在线用户
func SetOnlineUser(userId string) {
	defer func() {
		errs := recover()
		if errs != nil {
			logs.Loggers.Error("SetOnlineUser:", zap.Reflect("err", errs))
		}
	}()
	userStatusWg.Add(1)
	//性别在线
	go SetSexUserOnline(userId)
	//获取隐私状态
	var redisKey = constant.USER_PRIVACY_SET_KEY + userId
	conn := dao.RedisDefaultPool.Get()
	ret, _ := redis.String(conn.Do("get", redisKey))
	SetUserOnlineServer(userId)
	//如果是不存在隐私设置的 或者是公开的 才修改在线状态
	userPrivacySet := &mode.UserPrivacySet{}
	if ret != "" {
		userPrivacySet = mode.GetUserPrivacySet(ret)
	} else {
		userPrivacySet = nil
	}
	if userPrivacySet == nil || userPrivacySet.OnlineStatus == 1 {
		_, redisOlErr := conn.Do("hset", constant.USER_ONLINE_LIST, userId, time.Now().Unix())
		if redisOlErr != nil {
			logs.Loggers.Error("redisOlErr:", zap.Error(redisOlErr))
		}
		//设置es下线
		userStatusWg.Add(1)
		go UpdateEsOnLine(userId)

	}
	userStatusWg.Wait()
}

func GetUserIsOnline(userId string) int {
	defer func() {
		errs := recover()
		if errs != nil {
			logs.Loggers.Error("GetUserIsOnline:", zap.Reflect("err", errs))
		}
	}()
	conn := dao.RedisDefaultPool.Get()
	hexists, hexistsErr := redis.Int(conn.Do("HEXISTS", constant.USER_ONLINE_LIST, userId))
	if hexistsErr != nil {
		logs.Loggers.Error("hexistsErr:", zap.Error(hexistsErr))
	}
	return hexists
}

/***
设置用户在哪台服务器连接
*/
func SetUserOnlineServer(userId string) {
	conn := dao.RedisDefaultPool.Get()
	conn.Do("set", constant.USER_ONLINE_SERVER+userId, ServerNameKey)
}

/****
删除用户在哪台服务器连接
*/
func DelUserOnlineServer(userId string) {
	conn := dao.RedisDefaultPool.Get()
	conn.Do("DEL", constant.USER_ONLINE_SERVER+userId, ServerNameKey)
}

/*
/***
获取用户在哪台服务器登录

func GetUserOnlineServer(userId string) string {
	conn := dao.RedisDefaultPool.Get()
	getServerKey, getServerKeyErr := redis.String(conn.Do("get", constant.USER_ONLINE_SERVER+userId))
	if getServerKeyErr != nil {
		logs.Loggers.Error("GetUserOnlineServer-getServerKeyErr", zap.Error(getServerKeyErr))
	}
	if getServerKey != "" {
		domain, getDomainErr := redis.String(conn.Do("hget", getServerKey, "domain"))
		if getDomainErr != nil {
			logs.Loggers.Error("GetUserOnlineServer-getHostErr ", zap.Error(getServerKeyErr))
		}
		port, getPortErr := redis.String(conn.Do("hget", getServerKey, "port"))
		if getPortErr != nil {
			logs.Loggers.Error("GetUserOnlineServer-getPortErr", zap.Error(getServerKeyErr))
		}
		if domain != "" && port != "" {
			return domain + ":" + port
		}
	}
	return ""
}*/

//检查redis上的在线用户
func CheckRedisOnline() {
	defer func() {
		errs := recover()
		if errs != nil {
			logs.Loggers.Error("CheckRedisOnline:", zap.Reflect("err", errs))
		}
	}()
	logs.Loggers.Info("checkRedisOnline")
	//设置redis
	conn := dao.RedisDefaultPool.Get()
	redisOnlineMap, redisOlErr := redis.StringMap(conn.Do("HGETALL", constant.USER_ONLINE_LIST))
	if redisOlErr != nil {
		logs.Loggers.Error("redisOlErr:", zap.Error(redisOlErr))
	}
	logs.Loggers.Info("checkRedisOnline", zap.Any("redisOnlineMap", redisOnlineMap))
	nowTime := time.Now().Unix()
	for userId, userStrTime := range redisOnlineMap {
		getUserTime, timeErr := strconv.Atoi(userStrTime)
		if timeErr != nil {
			logs.Loggers.Error("timeErr", zap.Error(timeErr))
		}
		calculateTime := nowTime - int64(getUserTime)
		logs.Loggers.Info("CheckRedisOnline-calculateTime", zap.String(userId, strconv.Itoa(int(calculateTime))))
		if calculateTime > 50 {
			//性别离线
			//sexUserOffline(userId)
			SetOfflineUser(userId)
		}
	}
}

/***
根据用户性别 放到性别在线列表-上线
*/
func SetSexUserOnline(userId string) {
	userInfo := GetUserInfo(userId)
	if userInfo.Sex == 0 {
		return
	}
	redisKey := constant.USER_SEX_ONLINE_LIST + strconv.Itoa(userInfo.Sex)
	conn := dao.RedisDefaultPool.Get()
	conn.Do("HSET", redisKey, userId, time.Now().Unix())
	userStatusWg.Done()
}

/***
根据用户性别 放到性别离线列表-离线
*/
func SexUserOffline(userId string) {
	userInfo := GetUserInfo(userId)
	redisKey := constant.USER_SEX_ONLINE_LIST + strconv.Itoa(userInfo.Sex)
	conn := dao.RedisDefaultPool.Get()
	conn.Do("HDEL", redisKey, userId)
}

/****
根据性别获取 在线列表
*/
func GetSexOnline(sex string) []string {
	redisKey := constant.USER_SEX_ONLINE_LIST + sex
	conn := dao.RedisDefaultPool.Get()
	users, err := redis.Strings(conn.Do("HKEYS", redisKey))
	if err != nil {
		logs.Loggers.Error("GetUserSexOffline-users-err", zap.Error(err))
		return nil
	}
	return users
}

/***
获取在线用户
*/
func GetOnlineUser() []string {
	logs.Loggers.Info("GetOnlineUser")
	//设置redis
	conn := dao.RedisDefaultPool.Get()
	users, err := redis.Strings(conn.Do("HKEYS", constant.USER_ONLINE_LIST))
	if err != nil {
		logs.Loggers.Error("GetOnlineUser-HKEYS-err", zap.Error(err))
	}
	if users == nil {
		return nil
	}
	logs.Loggers.Info("获取在线用户", zap.Strings("GetOnlineUser", users))
	return users
}
