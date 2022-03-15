package constant

/****
路由编码
*/
const (
	//接收区

	ONLINE_CODE int32 = 1001 //上线

	SEND_MSG_CODE int32 = 1002 //发送普通消息

	SEND_CONTINUOUS_CODE int32 = 1003 //一键群发消息

	//响应区
	RESPONSE_HEARTBEAT_CODE int32 = 2001 //响应心跳

	RESPONSE_SEND_MSG_CODE int32 = 2002 //发送是否成功

	RESPONSE_GET_MSG_CODE int32 = 2003 //对方已收到消息

	RESPONSE_SYS_ERR_MSG_CODE int32 = 2004 //系统错误
)
