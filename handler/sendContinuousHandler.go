package handler

import (
	"github.com/fwhezfwhez/tcpx"
	"go.uber.org/zap"
	"im_socket_server/constant"
	"im_socket_server/logs"
	"im_socket_server/pb"
	"sync"
)

/***
1键群发用户
*/
var sendContinuousWg sync.WaitGroup

func SendContinuous(c *tcpx.Context) {
	var req pb.SendContinuousMsg
	_, eBindWithMarshaller := c.BindWithMarshaller(&req, tcpx.ProtobufMarshaller{})
	if eBindWithMarshaller != nil {
		logs.Loggers.Error("Online-eBindWithMarshaller", zap.Any("eBindWithMarshaller", eBindWithMarshaller))
		return
	}
	userIdLen := len(req.ToUserId)
	if userIdLen > 15 {
		var sysMsg pb.SysMsg
		sysMsg.Message = "人数超出范围"
		c.Reply(constant.RESPONSE_SEND_MSG_CODE, &sysMsg)
	}
	//声明一个存放userid的channel
	userIdChan := make(chan string, userIdLen)
	//声明一个主业务的处理channel
	pushMsgChan := make(chan string, userIdLen)
	//声明一个退出的chanel
	exitChan := make(chan bool, userIdLen)

	sendContinuousWg.Add(1)
	go addUserIdChan(userIdLen, userIdChan, req.ToUserId)
	//循环创建多协程来处理主业务
	for i := 0; i < userIdLen; i++ {
		sendContinuousWg.Add(1)
		go pushToUserMsg(c, userIdChan, pushMsgChan, exitChan, req.UserId,req.MsgContent)
		sendContinuousWg.Add(1)
		go sendIsOkToUser(c, pushMsgChan)
	}

	sendContinuousWg.Add(1)
	//结束业务的时候关闭 channel
	go func() {
		for i := 0; i < userIdLen; i++ {
			<-exitChan
		}
		//关闭主业务channel
		close(pushMsgChan)
		sendContinuousWg.Done()
	}()
	sendContinuousWg.Wait()
	//关闭退出
	close(exitChan)
}

/***
需要发送的userid
*/
func addUserIdChan(userIdLen int, userIdChan chan<- string, toUserId []string) {
	for i := 0; i < userIdLen; i++ {
		userIdChan <- toUserId[i]
	}
	close(userIdChan)
	sendContinuousWg.Done()
}

/****
处理主要业务
*/
func pushToUserMsg(c *tcpx.Context, userIdChan <-chan string, pushMsgChan chan<- string, exitChan chan<- bool, sendUserId, msgContent string) {
	for userId := range userIdChan {
		var getUserMsg pb.GetUserMsg
		getUserMsg.UserId = userId
		getUserMsg.SendUserId = sendUserId
		getUserMsg.MsgContent = msgContent
		//发送消息
		if c.GetPoolRef().GetClientPool(userId).IsOnline() {
			c.GetPoolRef().GetClientPool(userId).ProtoBuf(constant.RESPONSE_GET_MSG_CODE, &getUserMsg)
		} else {
			//自行保存未读消息
		}
		pushMsgChan <- userId
		exitChan <- true
	}
	sendContinuousWg.Done()
}

/***
处理次要的任务
*/
func sendIsOkToUser(c *tcpx.Context, pushMsgChan <-chan string) {
	for isSendOkToUser := range pushMsgChan {
		var sysMsg pb.SysMsg
		sysMsg.Message = isSendOkToUser + "ok"
		//响应发送成功
		c.Reply(constant.RESPONSE_SEND_MSG_CODE, &sysMsg)
	}
	sendContinuousWg.Done()

}
