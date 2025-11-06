package control

import (
	"StarHop/pb"

	"google.golang.org/protobuf/proto"
)

// TODO: 处理
func processRegisterPacket(data []byte) {
	// packet := parseRegisterPacket(data)
	_ = parseRegisterPacket(data)
	// TODO版本判断

}

// 解析注册包
func parseRegisterPacket(data []byte) *pb.RegisterPacket {
	r := &pb.RegisterPacket{}
	proto.Unmarshal(data, r)
	return r
}
