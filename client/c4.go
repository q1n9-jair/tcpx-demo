package main

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/fwhezfwhez/tcpx"
	"im_socket_server/constant"
	"im_socket_server/pb"
	"net"
	"sync"
	"time"
)

var wgTest sync.WaitGroup
func main() {
	conn, err := net.Dial("tcp", "localhost:8103")

	if err != nil {
		panic(err)
	}

	buf, e := tcpx.PackWithMarshallerName(tcpx.Message{
		MessageID: constant.ONLINE_CODE,
		Body: &pb.HeartBeat{
			UserId:  "1355156974324420610",
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

	var resp pb.SysMsg
	message, e := tcpx.UnpackWithMarshallerName(buf, &resp, "protobuf")
	if e != nil {
		panic(errorx.Wrap(e))
	}
	if e != nil {
		panic(errorx.Wrap(e))
	}
	fmt.Println("收到服务端消息块:", message.MessageID)
	fmt.Println("服务端消息:", resp.Message)
	for i:=0;i<5;i++ {
		wgTest.Add(1)
		go send()
	}

	
	//
	var heartBeat []byte
	heartBeat, e = tcpx.PackWithMarshallerName(tcpx.Message{
		MessageID: tcpx.DEFAULT_HEARTBEAT_MESSAGEID,
		Body: &pb.HeartBeat{
			UserId:  "1355156974324420610",
			AppCode: "callmi",
			Channel: "apple",
			/*BuildNumber: 45,
			PlatformType: "ios",*/
		},
	}, "protobuf")
	for i := 0; i < 5; i++ {
		_, e = conn.Write(heartBeat)
		if e != nil {
			fmt.Println(e.Error())
			break
		}
		time.Sleep(1 * time.Second)
	}
	wgTest.Wait()
}
func send()  {
	conn, err := net.Dial("tcp", "localhost:8103")
	if err != nil {
		panic(err)
	}
	toUserId:=make([]string,0)
	toUserId=append(toUserId,"1355156974324420610","1355156974324420610","1355156974324420610","1355156974324420610","1355156974324420610","1355156974324420610")
	buf, e := tcpx.PackWithMarshallerName(tcpx.Message{
		MessageID: constant.SEND_CONTINUOUS_CODE,
		Body: &pb.SendContinuousMsg{
			UserId:  "1355156974324420610",
			AppCode: "callmi",
			Channel: "apple",
			BuildNumber: 45,
			PlatformType: "ios",
			ToUserId: toUserId,
			MsgContent: "你好 ",
		},
	}, "protobuf")
	if e != nil {
		panic(e)
	}
	conn.Write(buf)
	wgTest.Done()
}
