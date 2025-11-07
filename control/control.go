package control

func InitControl() {
	// 初始化工作线程池
	// workers: 4 个工作线程
	// queueSize: 1000 个包的缓冲队列
	// TODO: 缓冲列队根据机器配置自适应
	initWorkerPool(10, 1000)
}

// 通道数据的接收
func receiveTunnelData(msg tunnelMsg) {
	// 对方的消息id
	_, ok := getMsgID(msg.data)
	if !ok {
		return
	}
	ptype, ok := getPacketType(msg.data)
	if !ok {
		return
	}
	switch ptype {
	case RegisterPacketType:
		processRegisterPacket(msg.id, msg.data[9:])
	}

}
