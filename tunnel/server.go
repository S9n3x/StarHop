package tunnel

import (
	"StarHop/control"
	"StarHop/pb"
	"StarHop/utils/logger"
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// hop Tunnel 实现
type hopTunnel struct {
	pb.UnimplementedHopTunnelServer
}

// 数据处理
func (t *hopTunnel) Stream(stream pb.HopTunnel_StreamServer) error {
	for {
		packet, err := stream.Recv()
		if err != nil {
			logger.Warn(err.Error())
			break
		}

		// 提交数据包
		control.SubmitPackage(packet.Data)
	}
	return nil
}

func Start() {
	cert, err := generateCert()
	if err != nil {
		logger.Error("failed to generate tunnel TLS certificate: ", err.Error())
	}
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	creds := credentials.NewTLS(tlsCfg)
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterHopTunnelServer(grpcServer, &hopTunnel{})

	grpcServer.Serve(getRandomListen())
}
