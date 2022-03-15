package handler

import (
	"fmt"
	"github.com/fwhezfwhez/tcpx"
	"go.uber.org/zap"
	"tcpx-demo/constant"
	"tcpx-demo/logs"
	"tcpx-demo/pb"
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
		getUserMsg := pb.GetUserMsg{
			UserId:     req.UserId,
			SendUserId: req.UserId,
			MsgContent: req.MsgContent,
		}
		//发送消息
		err = c.GetPoolRef().GetClientPool(req.ToUserId).ProtoBuf(constant.RESPONSE_GET_MSG_CODE, &getUserMsg)
		if err != nil {
			fmt.Println(err)
			c.Reply(constant.RESPONSE_SEND_MSG_CODE, &pb.SysMsg{Message: "fuck"})
			return
		}

	} else {
		//自行手写保存redis或mysql的未读消息
	}
	//响应发送成功
	c.Reply(constant.RESPONSE_SEND_MSG_CODE, &pb.SysMsg{Message: "ok"})
}
