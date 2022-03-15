package main

import (
	"fmt"
	"github.com/GUAIK-ORG/go-snowflake/snowflake"
	"github.com/fwhezfwhez/tcpx"
	"net"
	"strconv"
	"tcpx-demo/constant"
	"tcpx-demo/pb"
	"time"
)

func main() {
	for i := 0; i < 100; i++ {
		go func() {
			conn, err := net.Dial("tcp", "localhost:8103")
			if err != nil {
				panic(err)
			}
			s, err := snowflake.NewSnowflake(int64(0), int64(0))
			id := s.NextVal()
			uid := strconv.FormatInt(id, 10)
			touid := "123"
			online(conn, uid)

			ticker := time.NewTicker(1 * time.Second)
			for {
				select {
				case <-ticker.C:
					heartbeat(conn, uid)
					send(conn, uid, touid)
				}
			}

		}()
	}
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:

		}
	}

}

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

	var resp pb.SysMsg
	message, e := tcpx.UnpackWithMarshallerName(buf, &resp, "protobuf")
	if e != nil {
		fmt.Println(e)
	}

	fmt.Println("收到服务端消息块:", message.MessageID)
	fmt.Println("服务端消息:", resp.Message)
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
