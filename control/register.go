package control

import (
	"StarHop/pb"

	"github.com/golang/protobuf/proto"
)

type tunnelType int

const (
	grpcServer tunnelType = iota
	grpcClient
)

type tunnelStream interface {
	Send(*pb.HopPacket) error
	Recv() (*pb.HopPacket, error)
}
type tunnel struct {
	tType  tunnelType
	stream tunnelStream
}

// TODO: 处理
func processRegisterPacket(data []byte) {
	packet := parseRegisterPacket(data)
}

// 解析注册包
func parseRegisterPacket(data []byte) *pb.RegisterPacket {
	r := &pb.RegisterPacket{}
	proto.Unmarshal(data, r)
	return r
}
