package register

import (
	"StarHop/pb"
	"errors"
	"sort"
	"sync"
)

var (
	ErrRegistryFull    = errors.New("registry: connection limit reached")
	ErrNotFound        = errors.New("registry: connection not found")
	ErrNoConnAvailable = errors.New("registry: no connection available")
)

// tunnelStream 流接口封装
type tunnelStream interface {
	Send(*pb.HopPacket) error
	Recv() (*pb.HopPacket, error)
}

// TunnelConn 链接信息
type TunnelConn struct {
	Name     string       // 服务名字
	BackAddr string       // 回连地址
	Latency  int64        // 延迟
	Stream   tunnelStream // 流本体
}

// Send Conn直接调用发送数据
func (tc *TunnelConn) Send(pkt *pb.HopPacket) error {
	return tc.Stream.Send(pkt)
}

// Recv Conn直接调用接收数据
func (tc *TunnelConn) Recv() (*pb.HopPacket, error) {
	return tc.Stream.Recv()
}

// Registry 连接注册中心
type Registry struct {
	maxSize int                    // 连接数上限
	mu      sync.RWMutex           // 并发保护
	conns   map[string]*TunnelConn // Name -> Conn 快速查找
	sorted  []*TunnelConn          // 延迟排序
}

// NewRegistry 创建注册中心实例
func NewRegistry(maxSize int) *Registry {
	return &Registry{
		maxSize: maxSize,
		conns:   make(map[string]*TunnelConn),
		sorted:  make([]*TunnelConn, 0, maxSize),
	}
}

// Register 注册新连接到中心
func (r *Registry) Register(conn *TunnelConn) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.conns) >= r.maxSize {
		return ErrRegistryFull
	}

	if _, exists := r.conns[conn.Name]; exists {
		return errors.New("registry: connection ID already exists")
	}

	r.conns[conn.Name] = conn

	r.sorted = append(r.sorted, conn)

	r.sortLocked()

	return nil
}

// UpdateLatency 更新指定连接的延迟并重新排序
func (r *Registry) UpdateLatency(name string, latency int64) error {
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
func (r *Registry) GetBest() (*TunnelConn, error) {
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
func (r *Registry) Remove(name string) error {
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

// Count 返回当前连接数
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.conns)
}

// GetByID 根据 ID 获取连接
func (r *Registry) GetByID(id string) (*TunnelConn, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conn, exists := r.conns[id]
	if !exists {
		return nil, ErrNotFound
	}
	return conn, nil
}

// GetTunnelConnByName 获取指定名字的连接
func (r *Registry) GetTunnelConnByName(name string) *TunnelConn {
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
func (r *Registry) sortLocked() {
	sort.Slice(r.sorted, func(i, j int) bool {
		return r.sorted[i].Latency < r.sorted[j].Latency
	})
}
