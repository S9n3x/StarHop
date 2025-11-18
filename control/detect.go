package control

import (
	"StarHop/pb"
	"StarHop/tunnel/register"
	"StarHop/utils/general"
	"StarHop/utils/logger"
	"sync"
	"time"
)

// 延时探测管理
var (
	detectRegistryMu sync.Mutex
	detectRegistry   = make(map[uint64]detectEntry)
)

type detectEntry struct {
	name      string
	startedAt time.Time
}

func storeDetectEntry(id uint64, name string, ts time.Time) {
	detectRegistryMu.Lock()
	detectRegistry[id] = detectEntry{name: name, startedAt: ts}
	detectRegistryMu.Unlock()
}

func takeDetectEntry(id uint64) (string, time.Time, bool) {
	detectRegistryMu.Lock()
	entry, ok := detectRegistry[id]
	if ok {
		delete(detectRegistry, id)
	}
	detectRegistryMu.Unlock()
	return entry.name, entry.startedAt, ok
}

// 心跳检测
func StartHeartbeatDetection(name string) {
	// 30秒的心跳检测
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	for range ticker.C {
		pingID := NextMsgID()
		stream := register.Hub.GetTunnelConnByName(name)
		if stream == nil {
			logger.Warn("HeartbeatDetection: No stream found for ", name, " during heartbeat detection")
			if err := register.Hub.Remove(name); err != nil {
				logger.Warn("HeartbeatDetection: Failed to remove tunnel for ", name, ":", err.Error())
			}
			return
		}
		// 记录探测条目
		storeDetectEntry(pingID, name, time.Now())
		// 发送Ping包
		if err := stream.Send(&pb.HopPacket{
			Data: NewPacket(pingID, PingPacketType, []byte{1}),
		}); err != nil {
			logger.Warn("HeartbeatDetection: Failed to send Ping packet to ", name, ":", err.Error())
			if err := register.Hub.Remove(name); err != nil {
				logger.Warn("HeartbeatDetection: Failed to remove tunnel for ", name, ":", err.Error())
			}
			return
		}
	}

}

func processPingPacket(mid uint64, data []byte, kick chan struct{}) {
	// 获取数据包的id
	pid, ok := getMsgID(data)
	if !ok {
		logger.Warn("Failed to get packet ID from Ping packet")
		general.CloseStreamConn(kick)
		return
	}
	// 处理Ping包
	stream, _, err := register.PopWaitingMsg(mid)
	if err != nil {
		logger.Warn("Failed to get waiting message for register response:", err.Error())
		general.CloseStreamConn(kick)
		return
	}
	// 回复Pong包
	if err := stream.Send(&pb.HopPacket{
		Data: NewPacket(pid, PongPacketType, []byte{1}),
	}); err != nil {
		logger.Warn("Failed to send Pong packet:", err.Error())
		general.CloseStreamConn(kick)
		return
	}
}

func processPongPacket(mid uint64, data []byte, kick chan struct{}, time time.Time) {
	// 消息清除
	register.ClearWaitingMsg(mid)
	// 获取数据包的id
	pid, ok := getMsgID(data)
	if !ok {
		logger.Warn("Failed to get packet ID from Ping packet")
		general.CloseStreamConn(kick)
		return
	}
	name, stime, ok := takeDetectEntry(pid)
	if !ok {
		logger.Warn("Failed to take detect entry for packet ID:", pid)
		general.CloseStreamConn(kick)
		return
	}
	if err := register.Hub.UpdateLatency(name, time.Sub(stime).Milliseconds()); err != nil {
		logger.Warn("Failed to update latency for ", name, ":", err.Error())
	}
	// 输出延时
	logger.Debug("Latency for ", name, ": ", time.Sub(stime).Milliseconds(), " ms")
}
