package control

import (
	"StarHop/pb"
	"StarHop/tunnel/register"
	"StarHop/utils/logger"
	"StarHop/utils/meta"
	"strings"

	"google.golang.org/grpc/peer"
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

	// 获取客户端公网IP
	var clientIP string
	if serverStream, ok := stream.(pb.HopTunnel_StreamServer); ok {
		if p, ok := peer.FromContext(serverStream.Context()); ok {
			addr := p.Addr.String()
			if idx := strings.LastIndex(addr, ":"); idx != -1 {
				clientIP = addr[:idx]
			}
		}
	}

	backAddr := clientIP + ":" + packet.Port

	if err := register.Hub.Register(&register.TunnelConn{
		Name:     packet.Device,
		BackAddr: backAddr,
		Stream:   stream,
	}, true); err != nil {
		logger.Warn("Failed to register tunnel connection:", err.Error())
		return
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

	if err := stream.Send(&pb.HopPacket{Data: rData}); err != nil {
		register.Hub.Remove(packet.Device)
		logger.Warn("Failed to send register response:", err.Error())
		return
	}

}

// 解析注册包
func parseRegisterPacket(data []byte) *pb.RegisterPacket {
	r := &pb.RegisterPacket{}
	proto.Unmarshal(data, r)

	return r
}
