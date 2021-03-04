package main

import (
	"github.com/fwhezfwhez/tcpx"
	"github.com/robfig/cron"
	"go.uber.org/zap"
	config "im_socket_server/config"
	"im_socket_server/constant"
	"im_socket_server/handler"
	"im_socket_server/logs"
	"im_socket_server/pb"
	"im_socket_server/util"
	"log"
	"strconv"
	"time"
	//"tcpx"
)

type ConfigMode struct {
	Name    string `json:"name"`
	Host    string `json:"host"`
	TcpPort int    `json:"tcpPort"`
	Version string `json:"version"`
}

var imSocketServerLogo = `
              ██                        
              ▀▀                        
            ████        ▄█████▄▄▄███████▄           
              ██        ██▀   ██      ██   
              ██        ██    ██      ██    
           ▄▄▄██▄▄▄     ██    ██      ██    
           ▀▀▀▀▀▀▀▀     ▀▀    ▀▀      ▀▀ `
var configMode *ConfigMode

func main() {
	srv := tcpx.NewTcpX(tcpx.ProtobufMarshaller{})
	//开启自带用户在线池
	srv.WithBuiltInPool(true)
	srv.AddHandler(constant.ONLINE_CODE, handler.Online)
	srv.AddHandler(constant.SEND_MSG_CODE, handler.Send)
	srv.AddHandler(constant.SEND_CONTINUOUS_CODE, handler.SendContinuous)

	//自动检测掉线以及未发心跳
	srv.HeartBeatModeDetail(true, 20*time.Second, false, tcpx.DEFAULT_HEARTBEAT_MESSAGEID)
	//重写心跳
	srv.RewriteHeartBeatHandler(tcpx.DEFAULT_HEARTBEAT_MESSAGEID, func(c *tcpx.Context) {
		defer c.RecvHeartBeat()
		var req pb.HeartBeat
		_, tcpxErr := c.BindWithMarshaller(&req, tcpx.ProtobufMarshaller{})
		if tcpxErr != nil {
			logs.Loggers.Error("HeartBeat:", zap.Any("tcpxErr", tcpxErr))
			return
		}
		if req.UserId != "" {
			logs.Loggers.Info("HeartBeat:", zap.String("userId", req.UserId))
			//往redis续命
			go util.SetOnlineUser(req.UserId)
			//发送响应
			var rep pb.SysMsg
			rep.Message = "OnlineSuccess"
			eProtoBuf := c.Reply(constant.RESPONSE_HEARTBEAT_CODE, &rep)
			if eProtoBuf != nil {
				logs.Loggers.Error("HeartBeat:", zap.Any("eProtoBuf", eProtoBuf))
			}
		}
	})
	//检查下线
	srv.OnClose = func(c *tcpx.Context) {
		userId, _ := c.Username()
		if userId != "" {
			logs.Loggers.Info("Offline-自动检测掉线", zap.String("userId", userId))
			go util.SetOfflineUser(userId)
			pool := c.GetPoolRef().GetClientPool(userId)
			go util.SexUserOffline(userId)
			if pool != nil {
				pool.CloseConn()
				c.Offline()
			}
		}
	}
	//定时任务注册服务
	go Cron()
	//开始监听
	srv.ListenAndServe("tcp", configMode.Host+":"+strconv.Itoa(configMode.TcpPort))
}

/****
定时任务
*/
func Cron() {
	log.Println("Starting Cron...")
	c := cron.New()
	c.AddFunc("*/30 * * * * ?", util.ServerLives)
	c.AddFunc("*/25 * * * * ?", util.CheckRedisOnline)
	c.Start()
}

/***
加载配置
*/

func init() {
	log.Println(imSocketServerLogo)
	c := config.GetConfig()
	logs.Loggers.Info(c.GetString("host"))
	host := c.GetString("host")
	tcpPort := c.GetInt("tcpPort")
	version := c.GetString("version")

	mode := ConfigMode{}
	mode.Host = host
	mode.Name = util.GetExternalIP().String()
	mode.TcpPort = tcpPort
	mode.Version = version
	configMode = &mode
	logs.Loggers.Info("path:" + configMode.Host + ":" + strconv.Itoa(configMode.TcpPort))
	logs.Loggers.Info("version：" + configMode.Version)
	logs.Loggers.Info("serverName:" + configMode.Name)
	logs.Loggers.Info("------ log printl ----")
	//开始注册服务
	util.ServerRegistryCenter()
}
