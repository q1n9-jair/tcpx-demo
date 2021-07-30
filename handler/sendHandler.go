package handler

import (
	"github.com/fwhezfwhez/tcpx"
	"go.uber.org/zap"
	"im_socket_server/constant"
	"im_socket_server/logs"
	"im_socket_server/pb"
)

/***
发送消息业务处理
*/
func Send(c *tcpx.Context) {
	//接收
	var req pb.SendMsg
	_, err := c.BindWithMarshaller(&req, tcpx.ProtobufMarshaller{})
	if err != nil {
		logs.Loggers.Error("Online-BindWithMarshaller", zap.Error(err))
		return
	}
	//如果是多台机器请自行写一个查找用户在哪台机器上登录的 然后进行发送消息
	isOnline := c.GetPoolRef().GetClientPool(req.ToUserId).IsOnline()
	if isOnline {
		var getUserMsg pb.GetUserMsg
		getUserMsg.UserId = req.ToUserId
		getUserMsg.SendUserId = req.UserId
		getUserMsg.MsgContent = req.MsgContent
		//发送消息
		c.GetPoolRef().GetClientPool(req.ToUserId).ProtoBuf(constant.RESPONSE_GET_MSG_CODE, &getUserMsg)

	} else {
		//自行手写保存redis或mysql的未读消息
	}
	var sysMsg pb.SysMsg
	sysMsg.Message = "ok"
	//响应发送成功
	c.Reply(constant.RESPONSE_SEND_MSG_CODE, &sysMsg)
}
