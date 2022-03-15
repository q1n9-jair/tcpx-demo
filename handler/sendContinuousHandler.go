package handler

import (
	"context"
	"fmt"
	"github.com/fwhezfwhez/tcpx"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"tcpx-demo/constant"
	"tcpx-demo/logs"
	"tcpx-demo/pb"
)

/***
1键群发用户
*/

func SendContinuous(c *tcpx.Context) {
	var req pb.SendContinuousMsg
	_, eBindWithMarshaller := c.BindWithMarshaller(&req, tcpx.ProtobufMarshaller{})
	if eBindWithMarshaller != nil {
		logs.Loggers.Error("Online-eBindWithMarshaller", zap.Any("eBindWithMarshaller", eBindWithMarshaller))
		return
	}
	userIdLen := len(req.ToUserId)
	if userIdLen > 15 {
		c.Reply(constant.RESPONSE_SEND_MSG_CODE, &pb.SysMsg{Message: "人数超出范围"})
	}
	//声明一个存放userid的channel
	userIdChan := make(chan string, userIdLen)
	//声明一个主业务的处理channel
	pushMsgChan := make(chan string, userIdLen)
	//声明一个退出的chanel
	exitChan := make(chan bool, userIdLen)

	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		addUserIdChan(userIdLen, userIdChan, req.ToUserId)
		return nil
	})

	//循环创建多协程来处理主业务
	for i := 0; i < userIdLen; i++ {
		g.Go(func() error {
			pushToUserMsg(c, userIdChan, pushMsgChan, exitChan, req.UserId, req.MsgContent)
			return nil
		})

		g.Go(func() error {
			sendIsOkToUser(c, pushMsgChan)
			return nil
		})
	}

	//结束业务的时候关闭 channel
	g.Go(func() error {
		for i := 0; i < userIdLen; i++ {
			<-exitChan
		}
		//关闭主业务channel
		close(pushMsgChan)
		return nil
	})
	if err := g.Wait(); err != nil {
		c.Reply(constant.RESPONSE_SYS_ERR_MSG_CODE, &pb.SysMsg{Message: "系统错误"})
		return
	}
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
}

/****
处理主要业务
*/
func pushToUserMsg(c *tcpx.Context, userIdChan <-chan string, pushMsgChan chan<- string, exitChan chan<- bool, sendUserId, msgContent string) {
	for userId := range userIdChan {
		getUserMsg := pb.GetUserMsg{
			UserId:     userId,
			SendUserId: sendUserId,
			MsgContent: msgContent,
		}

		//发送消息
		if c.GetPoolRef().GetClientPool(userId).IsOnline() {
			err := c.GetPoolRef().GetClientPool(userId).ProtoBuf(constant.RESPONSE_GET_MSG_CODE, &getUserMsg)
			fmt.Println(err)
		} else {
			//自行保存未读消息
		}
		pushMsgChan <- userId
		exitChan <- true
	}
}

/***
处理次要的任务
*/
func sendIsOkToUser(c *tcpx.Context, pushMsgChan <-chan string) {
	for isSendOkToUser := range pushMsgChan {
		//响应发送成功
		c.Reply(constant.RESPONSE_SEND_MSG_CODE, &pb.SysMsg{Message: isSendOkToUser + "ok"})
	}
}
