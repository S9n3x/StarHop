package control

import (
	"StarHop/pb"
	"StarHop/tunnel/register"
	"StarHop/utils/logger"
	"errors"
	"io"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleIncomingStream 统一处理客户端和服务端的接收循环
func HandleIncomingStream(stream pb.Stream) error {
	kick := make(chan struct{}, 1)
	recvCh := make(chan *pb.HopPacket)
	errCh := make(chan error, 1)

	// 专门的接收 goroutine
	go func() {
		defer close(recvCh)
		defer close(errCh)
		for {
			pkt, err := stream.Recv()
			if err != nil {
				errCh <- err
				continue
			}
			recvCh <- pkt
		}
	}()

	for {
		select {
		case <-kick:
			return errors.New("stream closed by kick")

		case err, ok := <-errCh:
			if !ok {
				return errors.New("recv goroutine exited")
			}
			if streamRecvErrorHandle(err, stream) {
				return errors.New("stream closed")
			}
			continue

		case packet, ok := <-recvCh:
			if !ok {
				return errors.New("recv channel closed")
			}
			id := NextMsgID()
			register.PutWaitingMsg(id, stream)
			SubmitPackage(id, packet.Data, kick)
		}
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
	if strings.Contains(err.Error(), "stream closed by kick") {
		if name, ok := register.Hub.RemoveByStream(stream); ok {
			logger.Warn("Stream closed by kick", " name=", name, " err=", err.Error())
		} else {
			logger.Warn("Stream closed by kick (unregistered)", " err=", err.Error())
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

func RemoveStreamError(stream pb.Stream) string {
	var errMsg string

	if name, ok := register.Hub.RemoveByStream(stream); ok {
		logger.Warn("Stream closed by kick", " name=", name, " err=", err.Error())
	} else {
		logger.Warn("Stream closed by kick (unregistered)", " err=", err.Error())
	}
	return errMsg
}
