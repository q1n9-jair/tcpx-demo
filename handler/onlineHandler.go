package handler

import (
	"github.com/fwhezfwhez/tcpx"
	"go.uber.org/zap"
	"im_socket_server/constant"
	"im_socket_server/logs"
	"im_socket_server/pb"
	"im_socket_server/util"
)

/***
上线业务处理
*/
func Online(c *tcpx.Context) {

	//接收
	var req pb.HeartBeat
	_, eBindWithMarshaller := c.BindWithMarshaller(&req, tcpx.ProtobufMarshaller{})
	if eBindWithMarshaller != nil {
		logs.Loggers.Error("Online-eBindWithMarshaller", zap.Any("eBindWithMarshaller", eBindWithMarshaller))
		return
	}
	if req.UserId != "" {
		logs.Loggers.Info("Online", zap.String("UserId", req.UserId))
		//上线
		c.Online(req.UserId)
		//设置redis和es
		util.SetOnlineUser(req.UserId)
		//发送响应
		var rep pb.SysMsg
		rep.Message = "OnlineSuccess"
		eProtoBuf := c.Reply(constant.RESPONSE_HEARTBEAT_CODE, &rep)
		if eProtoBuf != nil {
			logs.Loggers.Error("Online-eProtoBuf", zap.Any("eProtoBuf", eProtoBuf))
		}
		//省略拉未读消息。。。。
	}
}
