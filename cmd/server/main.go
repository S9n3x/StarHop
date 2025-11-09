package main

import (
	"StarHop/control"
	"StarHop/pb"
	"StarHop/tunnel/client"
	"StarHop/tunnel/entrance"
	"StarHop/tunnel/register"
	"StarHop/utils/general"
	"StarHop/utils/logger"
	"StarHop/utils/meta"
	"StarHop/utils/service"
	"os"
)

func main() {
	lis, port := service.GetRandomListen()
	meta.Info.Port = port
	meta.Info.DeviceID = general.GenerateDeviceID()
	logger.Info("StarHop Starting Version:", meta.Info.Version, " DeviceID:", meta.Info.DeviceID, " Port:", meta.Info.Port)
	control.Init()
	for i := 0; i < register.Hub.MaxActiveOutboundConns; i++ {
		go client.CreateClientConn()
	}
	if len(os.Args) > 1 {
		control.StoreCandidateNodes(&pb.HopNodePacket{
			Address: os.Args[1],
		})
	}
	entrance.Start(lis)
}
