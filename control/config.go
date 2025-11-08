package control

// 工作线程数量
var workerNum int = 10

// 工作队列大小
var workerQueueSize int = 1000

// 注册中心最大连接数
var registryMaxConn int = 10

func SetWorkerNum(num int) {
	workerNum = num
}
func SetWorkerQueueSize(size int) {
	workerQueueSize = size
}
func SetRegistryMaxConn(maxConn int) {
	registryMaxConn = maxConn
}
