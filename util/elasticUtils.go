package util

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
	"go.uber.org/zap"
	"im_socket_server/dao"
	"im_socket_server/logs"
)

/***
修改用户下线状态
*/
func UpdateEsOffUser(userId, nowUnix string) bool {
	isEsUserInfo := Gets(userId)
	if !isEsUserInfo {
		logs.Loggers.Info("UpdateEsOffUser", zap.Bool("isEsUserInfoIsNull:"+userId, isEsUserInfo))
		return true
	}
	res, err := dao.Client.Update().Index("users").Id(userId).Doc(map[string]interface{}{"isOnLine": 0, "lastLoginDate": nowUnix}).Do(context.Background())
	if err != nil {
		logs.Loggers.Error("UpdateEsOffUser-err:", zap.Error(err))
		return false
	}
	jsonStr, errJson := json.Marshal(res)
	if errJson != nil {
		logs.Loggers.Error("UpdateEsOffUser-errJson", zap.Error(errJson))
	}
	logs.Loggers.Info("UpdateEsOffUser-json", zap.String("res", string(jsonStr)))
	return true
}

/***
修改用户在线状态
*/
func UpdateEsOnLine(userId string) {
	isEsUserInfo := Gets(userId)
	defer func() {
		errs := recover()
		if errs != nil {
			logs.Loggers.Error("UpdateEsOnLine:", zap.Reflect("err", errs))
		}
	}()
	if !isEsUserInfo {
		logs.Loggers.Info("UpdateEsOnLine", zap.Bool("isEsUserInfoIsNull:"+userId, isEsUserInfo))
		return
	}
	query := elastic.NewTermsQuery("userId", userId)
	script := elastic.NewScript("ctx._source.isOnLine = 1")
	_, err := dao.Client.UpdateByQuery().Index("users").Script(script).Query(query).Do(context.Background())
	if err != nil {
		logs.Loggers.Error("UpdateEsOnLine:", zap.Error(err))
	}
}

//查找
func Gets(userId string) bool {
	defer func() {
		errs := recover()
		if errs != nil {
			logs.Loggers.Error("UpdateEsOnLine:", zap.Reflect("err", errs))
		}
	}()
	//通过id查找
	getUser, err := dao.Client.Get().Index("users").Id(userId).Do(context.Background())
	if err != nil {
		logs.Loggers.Error("Gets:", zap.Error(err))
	}
	if getUser == nil {
		logs.Loggers.Info("Gets", zap.String("没有完成个人资料es信息不存在", userId))
		return false
	}
	return true
}
