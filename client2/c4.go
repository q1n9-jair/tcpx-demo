package main

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/fwhezfwhez/tcpx"
	"net"
	"tcpx-demo/constant"
	"tcpx-demo/pb"
	"time"
)

var packx = tcpx.NewPackx(tcpx.ProtobufMarshaller{})

func main() {

	conn, err := net.Dial("tcp", "localhost:8103")
	if err != nil {
		panic(err)
	}
	uid := "123"
	online(conn, uid)
	//开启协程 接收消息
	go func() {
		for {
			buf, e := tcpx.FirstBlockOf(conn)
			if e != nil {
				//if e == io.EOF {
				//	break
				//}
				panic(errorx.Wrap(e))
			}
			//	fmt.Println(buf)
			var resp pb.GetUserMsg
			message, e := tcpx.UnpackWithMarshallerName(buf, &resp, "protobuf")
			if e != nil {
				panic(errorx.Wrap(e))
			}
			if message.MessageID != constant.RESPONSE_HEARTBEAT_CODE {
				//fmt.Println("收到服务端消息块:", message)
				fmt.Println("服务端消息: 发送人：" + resp.UserId + ",内容:" + resp.MsgContent)
			}

		}
	}()

	//发送消息
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			heartbeat(conn, uid)

		}
	}

}

func sendQ(conn net.Conn) {
	toUserId := make([]string, 0)
	toUserId = append(toUserId, "1355156974324420610", "1355156974324420610", "1355156974324420610", "1355156974324420610", "1355156974324420610", "1355156974324420610")
	buf, e := tcpx.PackWithMarshallerName(tcpx.Message{
		MessageID: constant.SEND_CONTINUOUS_CODE,
		Body: &pb.SendContinuousMsg{
			UserId:       "1355156974324420610",
			AppCode:      "callmi",
			Channel:      "apple",
			BuildNumber:  45,
			PlatformType: "ios",
			ToUserId:     toUserId,
			MsgContent:   "你好 ",
		},
	}, "protobuf")
	if e != nil {
		panic(e)
	}
	conn.Write(buf)
}

//heartbeat 心跳
func heartbeat(conn net.Conn, uid string) {
	buf, e := tcpx.PackWithMarshallerName(tcpx.Message{
		MessageID: tcpx.DEFAULT_HEARTBEAT_MESSAGEID,
		Body: &pb.HeartBeat{
			UserId:  uid,
			AppCode: "callmi",
			Channel: "apple",
			/*BuildNumber: 45,
			PlatformType: "ios",*/
		},
	}, "protobuf")
	if e != nil {
		panic(e)
	}
	conn.Write(buf)

}

//online 上线
func online(conn net.Conn, uid string) {
	buf, e := tcpx.PackWithMarshallerName(tcpx.Message{
		MessageID: constant.ONLINE_CODE,
		Body: &pb.HeartBeat{
			UserId:       uid,
			AppCode:      "callmi",
			Channel:      "apple",
			BuildNumber:  45,
			PlatformType: "ios",
		},
	}, "protobuf")
	if e != nil {
		panic(e)
	}
	conn.Write(buf)
}

//send发送普通消息
func send(conn net.Conn, uid, toUid string) {
	buf, e := tcpx.PackWithMarshallerName(tcpx.Message{
		MessageID: constant.SEND_MSG_CODE,
		Body: &pb.SendMsg{
			UserId:       uid,
			AppCode:      "callmi",
			Channel:      "apple",
			BuildNumber:  45,
			PlatformType: "ios",
			ToUserId:     toUid,
			MsgContent:   "你好 ",
		},
	}, "protobuf")
	if e != nil {
		panic(e)
	}
	conn.Write(buf)
}
