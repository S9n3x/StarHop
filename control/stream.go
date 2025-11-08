package control

import (
	"StarHop/pb"
	"StarHop/tunnel/register"
	"StarHop/utils/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleIncomingStream 统一处理客户端和服务端的接收循环
func HandleIncomingStream(stream pb.Stream) {
	for {
		packet, err := stream.Recv()
		if err != nil {
			if streamRecvErrorHandle(err, stream) {
				return // 关闭接收
			}
			continue // 意外错误，继续接收
		}

		id := NextMsgID()
		register.PutWaitingMsg(id, stream)
		SubmitPackage(id, packet.Data)
	}
}

// 返回是否需要关闭接收任务
func streamRecvErrorHandle(err error, stream pb.Stream) bool {
	st, ok := status.FromError(err)
	if ok && (st.Code() == codes.Canceled || st.Code() == codes.DeadlineExceeded) {
		if name, ok := register.Hub.RemoveByStream(stream); ok {
			logger.Warn("Stream closed (context canceled)", " name=", name, " err=", err.Error())
		} else {
			logger.Warn("Stream closed (unregistered, context canceled)", " err=", err.Error())
		}
		return true
	}
	logger.Warn("Stream error (unknown error)", " err=", err.Error())
	return false
}
