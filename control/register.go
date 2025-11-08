package control

import (
	"StarHop/pb"
	"StarHop/tunnel/register"
	"StarHop/utils/logger"
	"StarHop/utils/meta"
	"StarHop/utils/service"

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

	// 获取回连地址
	backAddr := service.GetClientIPFromStream(stream) + ":" + packet.Port

	if err := register.Hub.Register(&register.TunnelConn{
		Name:     packet.Device,
		BackAddr: backAddr,
		Stream:   stream,
	}, true); err != nil {
		logger.Warn("Failed to register tunnel connection:", err.Error())
		if SendHopNodeList(stream, false) != nil {
			logger.Warn("Failed to send hop node list:", err.Error())
		} else {
			logger.Info("Sent hop node list to", packet.Device)
		}
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

	if err := stream.Send(&pb.HopPacket{Data: NewPacket(id, RegisterSuccessType, rData)}); err != nil {
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

func processRegisterSuccessPacket(id uint64, data []byte) {
	stream, err := register.PopWaitingMsg(id)
	if err != nil {
		logger.Warn("Failed to get waiting message for register response:", err.Error())
		return
	}
	var resp pb.RegisterPacket
	if err := proto.Unmarshal(data, &resp); err != nil {
		logger.Warn("Failed to unmarshal register response:", err.Error())
		return
	}
	backAddr := service.GetClientIPFromStream(stream) + ":" + resp.Port

	if err := register.Hub.Register(&register.TunnelConn{
		Name:     resp.Device,
		BackAddr: backAddr,
		Stream:   stream,
		Version:  resp.Version,
	}, false); err != nil {
		logger.Warn("Failed to register tunnel connection:", err.Error())
		return
	}
}

// 发送注册信息的节点
// 一般拒绝注册请求时返回当前可注册的其他节点
// 或者客户端用来获取所有节点信息
func SendHopNodeList(stream pb.Stream, isNat bool) error {
	var data pb.HopNodeListPacket
	for name, addr := range register.Hub.GetAllNodes(isNat) {
		data.Nodes = append(data.Nodes, &pb.HopNodePacket{
			Device:  name,
			Address: addr,
		})
	}
	hData, err := proto.Marshal(&data)
	if err != nil {
		return err
	}
	return stream.Send(&pb.HopPacket{Data: NewPacket(NextMsgID(), DisconnectPacketType, hData)})
}
