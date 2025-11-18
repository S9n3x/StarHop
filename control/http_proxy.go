package control

import (
	"StarHop/pb"
	"StarHop/tunnel/register"
	"StarHop/utils/general"
	"StarHop/utils/logger"
)

func processHttpProxyRequestPacket(mid uint64, data []byte, kick chan struct{}) {
	// 获取对方的消息ID
	pid, ok := getMsgID(data)
	if !ok {
		logger.Warn("Failed to get packet ID from HttpProxyRequest packet")
		general.CloseStreamConn(kick)
		return
	}
	rhops, ok := getHttpProxyPacketRemainingHops(data[9:])
	if !ok {
		logger.Warn("Failed to get HttpProxyPacket num hops from HttpProxyRequest packet")
		general.CloseStreamConn(kick)
		return
	}
	if rhops == 0 {
		// TODO 说明此条消息已经不需要转发，是交给自己的请求，需要进行代理处理，进行返回发送
		return
	}
	// 如果还需要转发，先保存上一个节点的id
	if !register.SetWaitingMsgPid(mid, pid) {
		// 保存上一个节点id失败
		logger.Warn("Failed to set waiting message PID")
		return
	}
	// 重新组装数据包
	stream, err := register.Hub.GetBest()
	if err != nil {
		logger.Warn("Failed to get best stream for HttpProxyRequest packet:", err.Error())
		return
	}
	// 组装数据转发到下一个节点,然后结束
	if err := stream.Send(&pb.HopPacket{
		Data: NewPacket(mid, HttpProxyRequestPacketType, (&httpProxyPacket{
			RemainingHops: rhops - 1,
			TotalHops:     data[10],
			Data:          data[11:],
		}).ToBytes()),
	}); err != nil {
		logger.Warn("Failed to forward HttpProxyRequest packet:", err.Error())
		general.CloseStreamConn(kick)
		return
	}
}

func processHttpProxyResponsePacket(mid uint64, data []byte, kick chan struct{}) {
	// 先把自己的等待消息队列拿出来，他已经无意义了
	_, _, err := register.PopWaitingMsg(mid)
	if err != nil {
		logger.Warn("Failed to pop waiting message:", err.Error())
		general.CloseStreamConn(kick)
		return
	}
	// 获取对方发送回来的消息id，这个id是之前转发过去的时候存进去的
	id, ok := getMsgID(data)
	if !ok {
		logger.Warn("Failed to get packet ID from HttpProxyResponse packet")
		general.CloseStreamConn(kick)
		return
	}
	// 判断自己是不是数据包的重点
	if needForward, err := needHttpProxyPacketForwarding(data[9:]); err != nil {
		logger.Warn("Failed to determine if HttpProxyResponse packet needs forwarding:", err.Error())
		general.CloseStreamConn(kick)
		return
	} else if !needForward {
		// TODO 说明已经到达最终节点,返回给代理客户端
	}
	// 获取跳点数量
	rhops, ok := getHttpProxyPacketRemainingHops(data[9:])
	if !ok {
		logger.Warn("Failed to get HttpProxyPacket num hops from HttpProxyRequest packet")
		general.CloseStreamConn(kick)
		return
	}

	// 取出之前放置过来的stream
	stream, pid, err := register.PopWaitingMsg(id)
	if err != nil {
		logger.Warn("Failed to get waiting message for HttpProxyResponse packet:", err.Error())
		return
	}
	if err := stream.Send(&pb.HopPacket{
		Data: NewPacket(pid, HttpProxyResponsePacketType, (&httpProxyPacket{
			RemainingHops: rhops - 1,
			TotalHops:     data[10],
			Data:          data[11:],
		}).ToBytes()),
	}); err != nil {
		logger.Warn("Failed to forward HttpProxyResponse packet:", err.Error())
		general.CloseStreamConn(kick)
		return
	}
}
