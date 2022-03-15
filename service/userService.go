package service

import (
	"encoding/json"
	"github.com/fwhezfwhez/tcpx/all-language-clients/model"
	"github.com/garyburd/redigo/redis"
	"go.uber.org/zap"
	"strconv"
	"tcpx-demo/constant"
	"tcpx-demo/dao"
	"tcpx-demo/logs"
	mode "tcpx-demo/model"
	"time"
)

/***
处理用户相关的操作业务
*/

type UserServices struct {
	//可放其他Services
}
type UserService interface {
	//SetOfflineUser 用户下线后处理的业务
	SetOfflineUser(userId string) error
	//SetOnlineUser 操作用户在线的业务
	SetOnlineUser(userId string) error
	//GetUserIsOnline 查看这个用户是否在线
	GetUserIsOnline(userId string) int
	//CheckRedisOnline 检查用户是否真实连接
	CheckRedisOnline()
	//GetUserInfo 获取用户基本信息
	GetUserInfo(userId string) (model.User, error)
	// SexUserOffline 根据用户性别 放到性别离线列表-离线
	SexUserOffline(userId string) error
	// SetUserOnlineServer 设置用户在哪台服务器连接
	SetUserOnlineServer(userId string) error
	// DelUserOnlineServer 删除用户在哪台服务器连接
	DelUserOnlineServer(userId string) error
}

func (u *UserServices) SetOfflineUser(userId string) error {
	//设置redis
	conn := dao.RedisDefaultPool.Get()
	//获取这个人是否为机器人
	isBotUser, err := redis.Bool(conn.Do("SISMEMBER", constant.BOT_USER_LIST, userId))
	if isBotUser {
		return nil
	}
	if err != nil {
		return err
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

	//如果为空或者等于1 就可以设置
	if userPrivacySet == nil || userPrivacySet.OnlineStatus == 1 {
		err = u.DelUserOnlineServer(userId)
		if err != nil {
			return err
		}
		//在redis上设置最后登录时间
		_, err = conn.Do("HSET", constant.USER_LAST_LOGIN_DATE, userId, nowUnix)
		if err != nil {
			return err
		}
		_, err = conn.Do("HDEL", constant.USER_ONLINE_LIST, userId)
		if err != nil {
			return err
		}
	}
	return nil
}

//设置在线用户
func (u *UserServices) SetOnlineUser(userId string) error {
	//性别在线
	err := u.SetSexUserOnline(userId)
	if err != nil {
		return err
	}
	//获取隐私状态
	var redisKey = constant.USER_PRIVACY_SET_KEY + userId
	conn := dao.RedisDefaultPool.Get()
	ret, _ := redis.String(conn.Do("get", redisKey))
	//如果是不存在隐私设置的 或者是公开的 才修改在线状态
	userPrivacySet := &mode.UserPrivacySet{}
	if ret != "" {
		userPrivacySet = mode.GetUserPrivacySet(ret)
	} else {
		userPrivacySet = nil
	}
	if userPrivacySet == nil || userPrivacySet.OnlineStatus == 1 {
		_, err = conn.Do("hset", constant.USER_ONLINE_LIST, userId, time.Now().Unix())
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *UserServices) GetUserIsOnline(userId string) int {
	conn := dao.RedisDefaultPool.Get()
	hexists, hexistsErr := redis.Int(conn.Do("HEXISTS", constant.USER_ONLINE_LIST, userId))
	if hexistsErr != nil {
		logs.Loggers.Error("service:userService:GetUserIsOnline:hexistsErr:", zap.Error(hexistsErr))
	}
	return hexists
}

//检查redis上的在线用户
func (u *UserServices) CheckRedisOnline() {
	//设置redis
	conn := dao.RedisDefaultPool.Get()
	redisOnlineMap, redisOlErr := redis.StringMap(conn.Do("HGETALL", constant.USER_ONLINE_LIST))
	if redisOlErr != nil {
		logs.Loggers.Error("service:userService:CheckRedisOnline:redisOlErr:", zap.Error(redisOlErr))
	}
	nowTime := time.Now().Unix()
	for userId, userStrTime := range redisOnlineMap {
		getUserTime, timeErr := strconv.Atoi(userStrTime)
		if timeErr != nil {
			logs.Loggers.Error("service:userService:CheckRedisOnline:timeErr", zap.Error(timeErr))
		}
		calculateTime := nowTime - int64(getUserTime)
		if calculateTime > 50 {
			//性别离线
			u.SexUserOffline(userId)
			u.SetOfflineUser(userId)
		}
	}
}

/***
根据用户性别 放到性别在线列表-上线
*/
func (u *UserServices) SetSexUserOnline(userId string) error {
	userInfo, err := u.GetUserInfo(userId)
	if userInfo.Sex == 0 {
		return nil
	}
	redisKey := constant.USER_SEX_ONLINE_LIST + strconv.Itoa(userInfo.Sex)
	conn := dao.RedisDefaultPool.Get()
	_, err = conn.Do("HSET", redisKey, userId, time.Now().Unix())
	if err != nil {
		return err
	}
	return nil
}

/***
获取用户基本信息
*/
func (u *UserServices) GetUserInfo(userId string) (mode.User, error) {
	conn := dao.RedisDefaultPool.Get()
	userInfo := &mode.User{}
	redisJson, err := redis.String(conn.Do("get", constant.USER_INFO+userId))
	if err != nil {
		return *userInfo, err
	}
	json.Unmarshal([]byte(redisJson), userInfo)
	return *userInfo, nil
}

/***
根据用户性别 放到性别离线列表-离线
*/
func (u *UserServices) SexUserOffline(userId string) error {
	userInfo, err := u.GetUserInfo(userId)
	redisKey := constant.USER_SEX_ONLINE_LIST + strconv.Itoa(userInfo.Sex)
	conn := dao.RedisDefaultPool.Get()
	_, err = conn.Do("HDEL", redisKey, userId)
	if err != nil {
		return err
	}
	return nil
}

/***
设置用户在哪台服务器连接
*/
func (u *UserServices) SetUserOnlineServer(userId string) error {
	conn := dao.RedisDefaultPool.Get()
	_, err := conn.Do("set", constant.USER_ONLINE_SERVER+userId, constant.ServerNameKey)
	return err
}

/****
删除用户在哪台服务器连接
*/
func (u *UserServices) DelUserOnlineServer(userId string) error {
	conn := dao.RedisDefaultPool.Get()
	_, err := conn.Do("DEL", constant.USER_ONLINE_SERVER+userId, constant.ServerNameKey)
	return err
}
