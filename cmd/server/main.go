package main

import (
	"StarHop/control"
	"StarHop/tunnel/client"
	"StarHop/tunnel/entrance"
	"StarHop/tunnel/register"
	"StarHop/utils/general"
	"StarHop/utils/logger"
	"StarHop/utils/meta"
	"StarHop/utils/service"
	"os"
	"time"
)

func main() {
	control.InitControl()
	register.StartAutoCleanup(1*time.Minute, 5*time.Minute)
	lis, port := service.GetRandomListen()
	meta.Info.Port = port
	meta.Info.DeviceID = general.GenerateDeviceID()
	logger.Info("StarHop Starting Version:", meta.Info.Version, " DeviceID:", meta.Info.DeviceID, " Port:", meta.Info.Port)
	if len(os.Args) > 1 {
		client.Register(os.Args[1])
	}
	entrance.Start(lis)
}
