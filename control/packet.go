package control

import (
	"encoding/binary"
	"sync/atomic"
)

type packetType uint8

const (
	// 通道注册
	// 这个类型使用RegisterPacket
	RegisterPacketType packetType = iota
	// 注册成功的返回状态
	RegisterSuccessType
	// 主动断开链接
	// 发生在剔除高延迟节点阶段，主节点链接数量已满，但是新来了一个低延迟的节点，给高延迟节点发送断开信号
	// 这个类型需要携带非NAT节点的地址信息HopNodeListPacket
	DisconnectPacketType
	// 延时探测
	PingPacketType
	// 延时探测回应
	PongPacketType

	// 合法包类型的最大值
	maxPacketType
)

// 自身消息ID
var msgIDCounter uint64

func NextMsgID() uint64 {
	return atomic.AddUint64(&msgIDCounter, 1)
}

func NewPacket(id uint64, ptype packetType, data []byte) []byte {
	return (&packet{
		ID:   id,
		Type: ptype,
		Data: data,
	}).toBytes()
}

type packet struct {
	ID   uint64
	Type packetType
	Data []byte
}

// packet转byte
func (p *packet) toBytes() []byte {
	buf := make([]byte, 9+len(p.Data)) // ID(8) + Type(1)
	binary.BigEndian.PutUint64(buf[0:8], p.ID)
	buf[8] = uint8(p.Type)
	copy(buf[9:], p.Data)
	return buf
}

// 提取packetType
func getPacketType(data []byte) (packetType, bool) {
	if len(data) < 9 {
		return 0, false
	}

	ptype := packetType(data[8])

	// 合法检查
	if ptype >= maxPacketType {
		return 0, false
	}

	return ptype, true
}

// 提取消息ID
func getMsgID(data []byte) (uint64, bool) {
	if len(data) < 9 {
		return 0, false
	}

	id := binary.BigEndian.Uint64(data[0:8])
	return id, true
}
