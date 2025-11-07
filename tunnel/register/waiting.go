package register

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrMsgNotFound = errors.New("registry: waiting message not found")
)

var (
	WaitingConns   = make(map[uint64]*waitingMsg)
	waitingConnsMu sync.RWMutex
)

type waitingMsg struct {
	Conn      *TunnelConn
	CreatedAt time.Time
}

// PopWaitingConn 取出链接
func PopWaitingConn(msgID uint64) (*TunnelConn, error) {
	waitingConnsMu.Lock()
	defer waitingConnsMu.Unlock()

	msg, exists := WaitingConns[msgID]
	if !exists {
		return nil, ErrMsgNotFound
	}

	delete(WaitingConns, msgID)
	return msg.Conn, nil
}

// PutWaitingConn 传入链接
func PutWaitingConn(msgID uint64, conn *TunnelConn) {
	waitingConnsMu.Lock()
	defer waitingConnsMu.Unlock()

	WaitingConns[msgID] = &waitingMsg{
		Conn:      conn,
		CreatedAt: time.Now(),
	}
}

// StartAutoCleanup 自动清理超时链接
func StartAutoCleanup(interval, timeout time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			waitingConnsMu.Lock()
			now := time.Now()
			for msgID, msg := range WaitingConns {
				if now.Sub(msg.CreatedAt) > timeout {
					delete(WaitingConns, msgID)
				}
			}
			waitingConnsMu.Unlock()
		}
	}()
}
