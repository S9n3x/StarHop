package test

import (
	"StarHop/control"
	"StarHop/tunnel/entrance"
	"StarHop/utils/logger"
	"StarHop/utils/meta"
	"testing"
)

func TestServer(t *testing.T) {
	// 启动主节点
	lis, port := getListen(":12345")
	meta.Info.Port = port

	meta.Info.DeviceID = "测试主节点"
	// 最小为2，因为至少要有1个非NAT节点，否则其他节点将无法通过他加入网络
	control.SetRegistryMaxConn(2) // 设置最大连接数为2，便于测试

	logger.Info("StarHop Starting Version:", meta.Info.Version, " DeviceID:", meta.Info.DeviceID, " Port:", meta.Info.Port)
	control.Init()

	entrance.Start(lis)
}
func TestClient1(t *testing.T) {
	createTestClient(1, ":12345", true)
}

func TestClient2(t *testing.T) {
	createTestClient(2, ":12345", false)
}

func TestClient3(t *testing.T) {
	createTestClient(3, ":12345", false)
}
