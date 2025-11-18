package control

import (
	"encoding/binary"
	"errors"
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

	// HTTP代理请求包
	HttpProxyRequestPacketType
	// HTTP代理返回包
	HttpProxyResponsePacketType

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

// 0-7是id
// 8是type
// 9:-数据
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

type httpProxyPacket struct {
	// 剩余需要转发节点的总数
	RemainingHops uint8
	// 一共需要转发的总数
	TotalHops uint8
	// 数据
	Data []byte
}

// 转bytes
func (p *httpProxyPacket) ToBytes() []byte {
	result := make([]byte, 2+len(p.Data))
	result[0] = p.RemainingHops
	result[1] = p.TotalHops

	copy(result[2:], p.Data)
	return result
}

// 获取HttpProxyPacket中的跳点数量
func getHttpProxyPacketRemainingHops(data []byte) (uint8, bool) {
	if len(data) < 1 {
		return 0, false
	}

	numHops := data[0]
	return numHops, true
}

// 判断是否需要继续回传节点
// true 继续转发
// flase 说明已经到达最终节点
func needHttpProxyPacketForwarding(data []byte) (bool, error) {
	if len(data) < 2 {
		// 数据不对
		return true, errors.New("needHttpProxyPacketForwarding data length less than 2")
	}
	if data[0] == data[1] {
		// 到达最终节点
		return false, nil
	}
	if data[0] > data[1] {
		// 数据出现错误，
		return true, errors.New("needHttpProxyPacketForwarding data error: remaining hops greater than total hops")
	}
	return true, nil
}

// 获取HttpProxyPacket中总跳点数量
func getHttpProxyPacketTotalHops(data []byte) (uint8, bool) {
	if len(data) < 2 {
		return 0, false
	}
	numHops := data[1]
	return numHops, true
}
