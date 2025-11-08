package register

import (
	"StarHop/pb"
	"StarHop/utils/logger"
	"errors"
	"net"
	"sort"
	"sync"
	"time"
)

var (
	ErrregistryFull    = errors.New("registry: connection limit reached")
	ErrNotFound        = errors.New("registry: connection not found")
	ErrNoConnAvailable = errors.New("registry: no connection available")
)

// TunnelConn 链接信息
type TunnelConn struct {
	Name     string    // 服务名字
	BackAddr string    // 回连地址
	Latency  int64     // 延迟
	Stream   pb.Stream // 流本体
	IsNAT    bool      // 是否为可回连
	Version  string    // 客户端版本
}

// Send Conn直接调用发送数据
func (tc *TunnelConn) Send(pkt *pb.HopPacket) error {
	return tc.Stream.Send(pkt)
}

// Recv Conn直接调用接收数据
func (tc *TunnelConn) Recv() (*pb.HopPacket, error) {
	return tc.Stream.Recv()
}

// registry 连接注册中心
type registryHub struct {
	maxSize int                    // 连接数上限
	mu      sync.RWMutex           // 并发保护
	conns   map[string]*TunnelConn // Name -> Conn 快速查找
	sorted  []*TunnelConn          // 延迟排序
}

// CreateRegistryHub 创建注册中心实例
func CreateRegistryHub(maxSize int) {
	Hub = registryHub{
		maxSize: maxSize,
		conns:   make(map[string]*TunnelConn),
		sorted:  make([]*TunnelConn, 0, maxSize),
	}
}

var Hub = registryHub{}

// Register 注册新连接到中心
func (r *registryHub) Register(conn *TunnelConn, test bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.conns) >= r.maxSize {
		logger.Debug(len(r.conns), "", r.maxSize)
		return ErrregistryFull
	}

	if _, exists := r.conns[conn.Name]; exists {
		return errors.New("registry: connection name already exists")
	}

	if test {
		conn.IsNAT = testNAT(conn.BackAddr)
	}

	r.conns[conn.Name] = conn

	r.sorted = append(r.sorted, conn)

	r.sortLocked()

	logger.Info("Device Registered -", " Device:", conn.Name, " addr:", conn.BackAddr, " IsNAT:", conn.IsNAT, " Version:", conn.Version)

	return nil
}

// 测试是否为nat
func testNAT(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return true
	}
	conn.Close()
	return false
}

// UpdateLatency 更新指定连接的延迟并重新排序
func (r *registryHub) UpdateLatency(name string, latency int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, exists := r.conns[name]
	if !exists {
		return ErrNotFound
	}

	conn.Latency = latency

	r.sortLocked()

	return nil
}

// GetBest 获取延迟最低的连接并移到队尾
func (r *registryHub) GetBest() (*TunnelConn, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.sorted) == 0 {
		return nil, ErrNoConnAvailable
	}

	bestConn := r.sorted[0]

	// 整体前移
	copy(r.sorted[0:], r.sorted[1:])

	// 把第一个放到最后
	r.sorted[len(r.sorted)-1] = bestConn

	return bestConn, nil
}

// Remove 移除指定连接
func (r *registryHub) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, exists := r.conns[name]
	if !exists {
		return ErrNotFound
	}

	delete(r.conns, name)

	for i, c := range r.sorted {
		if c.Name == name {
			r.sorted = append(r.sorted[:i], r.sorted[i+1:]...)
			break
		}
	}

	_ = conn
	return nil
}

// RemoveByStream 根据流进行删除
func (r *registryHub) RemoveByStream(s pb.Stream) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var removedName string
	for name, conn := range r.conns {
		if conn.Stream == s {
			// map和排序列表中都删除
			removedName = name
			delete(r.conns, name)
			for i, c := range r.sorted {
				if c == conn {
					r.sorted = append(r.sorted[:i], r.sorted[i+1:]...)
					break
				}
			}
			break
		}
	}
	if removedName == "" {
		return "", false
	}
	return removedName, true
}

// Count 返回当前连接数
func (r *registryHub) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.conns)
}

// GetByname 根据 name 获取连接
func (r *registryHub) GetByname(name string) (*TunnelConn, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conn, exists := r.conns[name]
	if !exists {
		return nil, ErrNotFound
	}
	return conn, nil
}

// GetTunnelConnByName 获取指定名字的连接
func (r *registryHub) GetTunnelConnByName(name string) *TunnelConn {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, conn := range r.sorted {
		if conn.Name == name {
			return conn
		}
	}
	return nil
}

// sortLocked 根据延迟排序
func (r *registryHub) sortLocked() {
	sort.Slice(r.sorted, func(i, j int) bool {
		return r.sorted[i].Latency < r.sorted[j].Latency
	})
}

// 获取所有连接名字列表
func (r *registryHub) ListAll() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.conns))
	for name := range r.conns {
		names = append(names, name)
	}
	return names
}
