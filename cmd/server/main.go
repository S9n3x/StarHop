package main

import (
	"StarHop/tunnel/entrance"
	"StarHop/tunnel/register"
	"StarHop/utils/general"
	"StarHop/utils/logger"
	"StarHop/utils/meta"
	"StarHop/utils/service"
	"time"
)

func main() {
	register.StartAutoCleanup(1*time.Minute, 5*time.Minute)
	lis, port := service.GetRandomListen()
	meta.Info.Port = port
	meta.Info.DeviceID = general.GenerateDeviceID()
	logger.Info("StarHop Starting Version:", meta.Info.Version, " DeviceID:", meta.Info.DeviceID, " Port:", meta.Info.Port)
	entrance.Start(lis)
}
