package control

import (
	"StarHop/tunnel/register"
	"time"
)

func Init() {
	// 初始化工作线程池
	initWorkerPool(workerNum, workerQueueSize)
	// 启动等待消息自动清理
	register.StartWaitingMsgsAutoCleanup(1*time.Minute, 5*time.Minute)
	// 创建注册中心，可管理的最大连接数
	register.CreateRegistryHub(registryMaxConn)
}

// 通道数据的接收
func receiveTunnelData(msg tunnelMsg) {

	ptype, ok := getPacketType(msg.data)
	if !ok {
		return
	}
	switch ptype {
	case RegisterPacketType:
		processRegisterPacket(msg.id, msg.data[9:], msg.kick)
	case RegisterSuccessType:
		processRegisterSuccessPacket(msg.id, msg.data[9:], msg.kick)
	case DisconnectPacketType:
		processDisconnectPacket(msg.id, msg.data[9:], msg.kick)
	default:
		// 未知包类型，丢弃
	}
}
