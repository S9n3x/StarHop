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
	// EOF 关闭
	if errors.Is(err, io.EOF) {
		logAndRemoveStream(stream, err, "Connect closed (EOF)")
		return true
	}
	// 主动关闭
	if strings.Contains(err.Error(), "stream closed by kick") {
		logAndRemoveStream(stream, err, "Connect closed by kick")
		return true
	}
	// grpc错误(不全，后续慢慢维护错误类型)
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Canceled, codes.DeadlineExceeded, codes.Unavailable:
			logAndRemoveStream(stream, err, "Connect closed ("+st.Code().String()+")")
			return true
		}
	}
	logger.Warn("Unknown stream error (not closing stream)", " err=", err.Error())
	return false
}

func logAndRemoveStream(stream pb.Stream, err error, reason string) {
	if name, ok := register.Hub.RemoveByStream(stream); ok {
		logger.Warn(reason,
			" name=", name,
			" err=", err.Error(),
		)
	} else {
		logger.Warn(reason+" (unregistered)",
			" err=", err.Error(),
		)
	}
}
