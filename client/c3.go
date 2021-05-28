package main

import (
	"fmt"
	"github.com/GUAIK-ORG/go-snowflake/snowflake"
	"github.com/fwhezfwhez/errorx"
	"github.com/fwhezfwhez/tcpx"
	"github.com/golang/glog"
	"im_socket_server/pb"
	"net"
	"runtime"
	"strconv"
	"sync"
)

var wg sync.WaitGroup

func main() {
	fmt.Println(runtime.NumCPU())
	wg.Add(1)
	go tskTcp()
	wg.Wait()

}

func tskTcp() {

	for kk := 1; kk < 20000; kk++ {
		s, err := snowflake.NewSnowflake(int64(0), int64(0))
		id := s.NextVal()
		fmt.Println(kk)
		if err != nil {
			glog.Error(err)
			return
		}
		conn, err := net.Dial("tcp", "127.0.0.1:8103")

		if err != nil {
			panic(err)
		}
		buf, e := tcpx.PackWithMarshallerName(tcpx.Message{
			MessageID: 1,
			Body: &pb.HeartBeat{
				UserId:  strconv.FormatInt(id, 10),
				AppCode: "callmi",
				Channel: "xiaomi",
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
		fmt.Println("收到服务端消息块:", message.MessageID)
		fmt.Println("服务端消息:", resp.Message)
		//
	}
	wg.Done()
}
