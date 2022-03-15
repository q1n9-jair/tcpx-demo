package model

import "encoding/json"

type UserPrivacySet struct {
	Id int64 `json:"id"`
	/**
	 * userId
	 */
	UserId string `json:"userId"`
	/**
	 * 是否展示在线状态 0不展示 1公开展示
	 */
	OnlineStatus int64 `json:"onlineStatus"`
	/**
	 * 是否展示距离 0不展示 1展示
	 */
	Distance int64 `json:"distance"`
	/**
	 * 是否语音聊天 0不展示 1展示
	 */
	VoiceChat int64 `json:"voiceChat"`
	/**
	 * 相册展示  0收费展示 1公开展示
	 */
	PhotoAlbum int64 `json:"photoAlbum"`
	/**
	 * 解锁相册费用
	 */
	PhotoMoney int64 `json:"photoMoney"`
	/**
	 * 主页展示 0付费展示  1公开
	 */
	HomePage int64 `json:"homePage"`
	/**
	 * 解锁主页费用编码
	 */
	HomeMoneyId int64 `json:"homeMoneyId"`
	/**
	 * 创建时间
	 */
	CreatedDate interface{} `json:"createdDate"`
	/**
	 * 更新时间
	 */
	UpdatedDate interface{} `json:"updatedDate"`
}

/***

 */
func GetUserPrivacySet(jsonStr string) *UserPrivacySet {
	userPrivacySet := UserPrivacySet{}
	json.Unmarshal([]byte(jsonStr), &userPrivacySet)
	return &userPrivacySet
}
