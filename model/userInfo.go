package model

type User struct {
	AppCode     string `json:"appCode"`
	CreatedDate int64  `json:"createdDate"`
	Empty       bool   `json:"empty"`
	FaceAuth    int32  `json:"faceAuth"`
	HeadImage   string `json:"headImage"`
	Id          string `json:"id"`
	ImPassword  string `json:"imPassword"`
	InviteCode  string `json:"inviteCode"`
	IsComplete  int32  `json:"isComplete"`
	IsLoveVip   int32  `json:"isLoveVip"`
	IsRobot     int32  `json:"isRobot"`
	IsVip       int32  `json:"isVip"`
	NickName    string `json:"nickName"`
	NotEmpty    bool   `json:"notEmpty"`
	RealAuth    int32  `json:"realAuth"`
	Sex         int    `json:"sex"`
	Status      int32  `json:"status"`
	Telephone   string `json:"telephone"`
	Uid         string `json:"uid"`
	UpdatedDate int64  `json:"updatedDate"`
	WxCode      string `json:"wxCode"`
}
