package test

import (
	"StarHop/control"
	"StarHop/tunnel/client"
	"StarHop/tunnel/entrance"
	"StarHop/utils/logger"
	"StarHop/utils/meta"
	"StarHop/utils/service"
	"fmt"
	"net"
	"time"
)

// 快速创建
// 创建本地的非NAT节点
func createTestClient(id int, port string, server bool) {

	meta.Info.DeviceID = fmt.Sprintf("测试节点-%d", id)
	control.Init()
	if server {
		lis, port := service.GetRandomListen()
		meta.Info.Port = port
		go entrance.Start(lis)
	}
	logger.Info("StarHop Starting Version:", meta.Info.Version, " DeviceID:", meta.Info.DeviceID, " Port:", meta.Info.Port)

	if port != "" {
		client.Register(fmt.Sprintf("127.0.0.1%s", port))
	}

	time.Sleep(10 * time.Minute)
}

func getListen(p string) (net.Listener, string) {
	lis, err := net.Listen("tcp", p)
	if err != nil {
		logger.Error("failed to listen on port", err.Error())
	}
	_, port, err := net.SplitHostPort(lis.Addr().String())
	if err != nil {
		logger.Error("failed to parse listener address: ", err.Error())
	}
	return lis, port
}
