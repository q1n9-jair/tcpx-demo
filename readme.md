## tcpx框架文档
- https://github.com/fwhezfwhez/tcpx/blob/master/README-CN.md
## 环境要求
必要依赖
- protoc
- proto-gen-go
- proto 3.10.0
- 运行环境必须是64位
- 设置代理GOPROXY=https://goproxy.cn
- im 即时通讯服务
## 环境配置

- linux
- export GOPROXY=https://goproxy.cn
- CGO_ENABLED=0
- GOARCH=amd64
- export GO111MODULE=on

- win
- go env -w GOPROXY=https://goproxy.io
- SET CGO_ENABLED=0
- SET GOARCH=amd64
- go env -w GO111MODULE= on

## 项目结构体：
- client(模拟客户端测试)
- config(配置文件以及读取配置文件)
- constant(系统常量)
- handler(处理业务handler)
- main(主程序 包含路由)
- mode(结构体)
- pb(proto)
- util(工具类)
