package handler

import (
	"github.com/fwhezfwhez/tcpx"
	"go.uber.org/zap"
	"im_socket_server/constant"
	"im_socket_server/logs"
	"im_socket_server/pb"
	"im_socket_server/service"
)

/***
上线业务处理
*/
func Online(c *tcpx.Context) {

	//接收
	var req pb.HeartBeat
	_, err := c.BindWithMarshaller(&req, tcpx.ProtobufMarshaller{})
	if err != nil {
		logs.Loggers.Error("handler:Online:BindWithMarshaller", zap.Error(err))
		return
	}
	if req.UserId != "" {
		logs.Loggers.Info("Online", zap.String("UserId", req.UserId))
		//上线
		c.Online(req.UserId)
		//设置redis和es
		userServices := service.UserServices{}
		err = userServices.SetOnlineUser(req.UserId)
		if err != nil {
			logs.Loggers.Error("handler:Online:SetOnlineUser", zap.Error(err))
			return
		}
		//发送响应
		var rep pb.SysMsg
		rep.Message = "OnlineSuccess"
		eProtoBuf := c.Reply(constant.RESPONSE_HEARTBEAT_CODE, &rep)
		if eProtoBuf != nil {
			logs.Loggers.Error("Online-eProtoBuf", zap.Any("eProtoBuf", eProtoBuf))
		}
		//省略拉未读消息的业务。。。。
	}
}
