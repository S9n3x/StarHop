package control

func InitControl() {
	// 初始化工作线程池
	// workers: 4 个工作线程
	// queueSize: 1000 个包的缓冲队列
	// TODO: 缓冲列队根据机器配置自适应
	initWorkerPool(10, 1000)
}

// 通道数据的接收
func receiveTunnelData(data []byte) {
	packetType, ok := getPacketType(data)
	if !ok {
		return
	}

	switch packetType {
	case registerPacketType:
		processRegisterPacket(data[9:])
	}
}
