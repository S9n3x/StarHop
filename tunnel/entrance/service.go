package entrance

import (
	"StarHop/control"
	"StarHop/pb"
	"StarHop/utils/logger"
	"StarHop/utils/service"
	"crypto/tls"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// hop Tunnel 实现
type hopTunnel struct {
	pb.UnimplementedHopTunnelServer
}

// 数据处理
func (t *hopTunnel) Stream(stream pb.HopTunnel_StreamServer) error {
	return control.HandleIncomingStream(stream)
}

func Start(lis net.Listener) {
	cert, err := service.GenerateCert()
	if err != nil {
		logger.Error("failed to generate tunnel TLS certificate: ", err.Error())
	}
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	creds := credentials.NewTLS(tlsCfg)
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterHopTunnelServer(grpcServer, &hopTunnel{})

	grpcServer.Serve(lis)
}
