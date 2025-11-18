package control

import (
	"StarHop/pb"
	"StarHop/tunnel/register"
	"StarHop/utils/general"
	"StarHop/utils/logger"
	"StarHop/utils/meta"
	"StarHop/utils/service"

	"google.golang.org/protobuf/proto"
)

// processRegisterPacket 注册包处理
func processRegisterPacket(mid uint64, data []byte, kick chan struct{}) {

	// packet := parseRegisterPacket(data)
	packet := parseRegisterPacket(data)
	// TODO版本判断

	stream, _, err := register.PopWaitingMsg(mid)
	if err != nil {
		logger.Warn("Failed to get waiting message for register response:", err.Error())
		general.CloseStreamConn(kick)
		return
	}

	// 获取回连地址
	backAddr := service.GetClientIPFromStream(stream) + ":" + packet.Port

	if err := register.Hub.Register(&register.TunnelConn{
		Name:     packet.Device,
		BackAddr: backAddr,
		Stream:   stream,
		Version:  packet.Version,
	}, true); err != nil {
		logger.Warn("Failed to register tunnel connection:", err.Error())
		if SendHopNodeList(stream, false) != nil {
			logger.Warn("Failed to send hop node list:", err.Error())
		} else {
			logger.Info("Sent hop node list to ", packet.Device)
		}
		general.CloseStreamConn(kick)
		return
	}
	// 启动心跳检测
	go StartHeartbeatDetection(packet.Device)

	rData, err := proto.Marshal(&pb.RegisterPacket{
		Device:  meta.Info.DeviceID,
		Port:    meta.Info.Port,
		Version: meta.Info.Version,
	})
	if err != nil {
		logger.Warn("Failed to marshal register packet:", err.Error())
		general.CloseStreamConn(kick)
		return
	}

	if err := stream.Send(&pb.HopPacket{Data: NewPacket(mid, RegisterSuccessType, rData)}); err != nil {
		register.Hub.Remove(packet.Device)
		logger.Warn("Failed to send register response:", err.Error())
		general.CloseStreamConn(kick)
		return
	}
}

// 解析注册包
func parseRegisterPacket(data []byte) *pb.RegisterPacket {
	r := &pb.RegisterPacket{}
	proto.Unmarshal(data, r)

	return r
}

// 注册成功后反向注册服务端
func processRegisterSuccessPacket(mid uint64, data []byte, kick chan struct{}) {
	stream, _, err := register.PopWaitingMsg(mid)
	if err != nil {
		logger.Warn("Failed to get waiting message for register response:", err.Error())
		return
	}
	var resp pb.RegisterPacket
	if err := proto.Unmarshal(data, &resp); err != nil {
		logger.Warn("Failed to unmarshal register response:", err.Error())
		return
	}
	addr, ok := TakeAddrForStream(stream)
	if !ok {
		logger.Warn("Failed to take address for stream during register success processing")
		general.CloseStreamConn(kick)
		return
	}
	if err := register.Hub.Register(&register.TunnelConn{
		Name:     resp.Device,
		BackAddr: addr,
		Stream:   stream,
		Version:  resp.Version,
	}, false); err != nil {
		logger.Warn("Failed to register tunnel connection:", err.Error())
		return
	}
	// 启动心跳检测
	go StartHeartbeatDetection(resp.Device)
}

// 处理断开连接包
func processDisconnectPacket(data []byte, kick chan struct{}) {
	var resp pb.HopNodeListPacket
	if err := proto.Unmarshal(data, &resp); err != nil {
		logger.Warn("Failed to unmarshal disconnect response:", err.Error())
		return
	}
	for _, node := range resp.Nodes {
		logger.Info("Node:", node.Device, " Address:", node.Address)
	}

	StoreCandidateNodes(resp.Nodes...)
	general.CloseStreamConn(kick)
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

// Host记录
var addrStreamMap = make(map[pb.Stream]string)

func StoreAddrForStream(stream pb.Stream, addr string) {
	addrStreamMap[stream] = addr
}

func TakeAddrForStream(stream pb.Stream) (string, bool) {
	addr, ok := addrStreamMap[stream]
	if ok {
		delete(addrStreamMap, stream)
	}
	return addr, ok
}
