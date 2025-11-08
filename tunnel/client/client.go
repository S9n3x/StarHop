package client

import (
	"StarHop/control"
	"StarHop/pb"
	"StarHop/tunnel/register"
	"StarHop/utils/logger"
	"StarHop/utils/meta"
	"context"
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
)

func Register(addr string) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	creds := credentials.NewTLS(tlsConfig)

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		logger.Warn("Unable to connect to the server:", err.Error())
		return
	}
	defer conn.Close()

	client := pb.NewHopTunnelClient(conn)
	stream, err := client.Stream(context.Background())
	if err != nil {
		logger.Warn("Unable to establish stream: ", err.Error())
		return
	}
	rData, err := proto.Marshal(&pb.RegisterPacket{
		Device:  meta.Info.DeviceID,
		Port:    meta.Info.Port,
		Version: meta.Info.Version,
	})
	if err != nil {
		logger.Warn("Failed to marshal register packet:", err.Error())
		return
	}
	id := control.NextMsgID()
	register.PutWaitingMsg(id, stream)
	stream.Send(&pb.HopPacket{Data: control.NewPacket(id, control.RegisterPacketType, rData)})

	// 无需处理错误，客户端会自动关闭链接
	// 无错误的时候会一直进行处理，出现错误函数结束运行
	control.HandleIncomingStream(stream)
}
