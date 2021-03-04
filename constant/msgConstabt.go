package constant

/****
路由编码
*/
const (
	//接收区
	//上线
	ONLINE_CODE int32 = 1001
	//发送普通消息
	SEND_MSG_CODE        int32 = 1002
	//一键群发消息
	SEND_CONTINUOUS_CODE int32 = 1003



	//响应区
	// 响应心跳
	RESPONSE_HEARTBEAT_CODE int32 = 2001
	//发送是否成功
	RESPONSE_SEND_MSG_CODE int32 =2002
	//对方已收到消息
	RESPONSE_GET_MSG_CODE int32 = 2003
)
