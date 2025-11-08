package control

import (
	"StarHop/pb"
	"StarHop/tunnel/register"
	"StarHop/utils/logger"
	"errors"
	"io"

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
	if errors.Is(err, io.EOF) {
		// 正常关闭
		if name, ok := register.Hub.RemoveByStream(stream); ok {
			logger.Warn("Stream closed, removed registered connection", " name=", name, " err=", err.Error())
		} else {
			logger.Warn("Stream closed (unregistered)", " err=", err.Error())
		}
		return true
	}

	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.Canceled, codes.DeadlineExceeded, codes.Unavailable:
			// Canceled/超时/不可用，都需要关闭流
			if name, ok := register.Hub.RemoveByStream(stream); ok {
				logger.Warn("Stream closed", " name=", name, " code=", st.Code().String(), " err=", err.Error())
			} else {
				logger.Warn("Stream closed (unregistered)", " code=", st.Code().String(), " err=", err.Error())
			}
			return true
		}
	}

	logger.Warn("Stream error (unknown error)", " err=", err.Error())
	return false
}
