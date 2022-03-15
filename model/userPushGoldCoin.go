package model

type UserPushGoldCoinMsg struct {
	UserId    string `json:"userId"`
	HeadImage string `json:"headImage"`
	NickName  string `json:"nickName"`
}
