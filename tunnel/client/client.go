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
	stream.Send(&pb.HopPacket{Data: control.NewPacket(control.NextMsgID(), control.RegisterPacketType, rData)})
	packet, err := stream.Recv()
	if err != nil {
		logger.Warn("Failed to receive register response:", err.Error())
		return
	}
	var resp pb.RegisterPacket
	if err := proto.Unmarshal(packet.Data, &resp); err != nil {
		logger.Warn("Failed to unmarshal register response:", err.Error())
		return
	}
	if err := register.Hub.Register(&register.TunnelConn{
		Name:     resp.Device,
		BackAddr: addr,
		Stream:   stream,
		Version:  resp.Version,
	}, false); err != nil {
		logger.Warn("Failed to register tunnel connection:", err.Error())
		return
	}

	control.HandleIncomingStream(stream)
}
