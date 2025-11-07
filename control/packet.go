package control

import (
	"encoding/binary"
	"sync/atomic"
)

type packetType uint8

const (
	// 通道注册
	RegisterPacketType packetType = iota

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

// byte转packet
func bytesToPacket(data []byte) *packet {
	if len(data) < 9 {
		return nil
	}
	return &packet{
		ID:   binary.BigEndian.Uint64(data[0:8]),
		Type: packetType(uint8(data[8])),
		Data: data[9:],
	}
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
