package control

import (
	"StarHop/pb"
	"StarHop/tunnel/register"
	"StarHop/utils/logger"
	"StarHop/utils/meta"

	"google.golang.org/protobuf/proto"
)

// processRegisterPacket 注册包处理
func processRegisterPacket(id uint64, data []byte) {

	// packet := parseRegisterPacket(data)
	packet := parseRegisterPacket(data)
	// TODO版本判断

	stream, err := register.PopWaitingMsg(id)
	if err != nil {
		logger.Warn("Failed to get waiting message for register response:", err.Error())
	}
	rData, err := proto.Marshal(&pb.RegisterPacket{
		Device:  meta.Info.DeviceID,
		Port:    meta.Info.Port,
		Version: meta.Info.Version,
	})
	if err != nil {
		logger.Warn("Failed to marshal register packet:", err.Error())
		return
	}
	stream.Send(&pb.HopPacket{Data: rData})
	//TODO 注册到注册中心

	logger.Info("Device Registered -", " Device:", packet.Device, " Port:", packet.Port, " Version:", packet.Version)
}

// 解析注册包
func parseRegisterPacket(data []byte) *pb.RegisterPacket {
	r := &pb.RegisterPacket{}
	proto.Unmarshal(data, r)

	return r
}
