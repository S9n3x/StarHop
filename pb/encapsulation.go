package pb

// Stream 流接对象一封装
type Stream interface {
	Send(*HopPacket) error
	Recv() (*HopPacket, error)
}
