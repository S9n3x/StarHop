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
	WaitingMsgs   = make(map[uint64]*waitingMsg)
	WaitingMsgsMu sync.RWMutex
)

type waitingMsg struct {
	Stream    tunnelStream
	CreatedAt time.Time
}

// PopWaitingMsg 取出消息
func PopWaitingMsg(msgID uint64) (tunnelStream, error) {
	WaitingMsgsMu.Lock()
	defer WaitingMsgsMu.Unlock()

	msg, exists := WaitingMsgs[msgID]
	if !exists {
		return nil, ErrMsgNotFound
	}

	delete(WaitingMsgs, msgID)
	return msg.Stream, nil
}

// PutWaitingMsg 传入链接
func PutWaitingMsg(msgID uint64, conn tunnelStream) {
	WaitingMsgsMu.Lock()
	defer WaitingMsgsMu.Unlock()

	WaitingMsgs[msgID] = &waitingMsg{
		Stream:    conn,
		CreatedAt: time.Now(),
	}
}

// StartAutoCleanup 自动清理超时链接
func StartAutoCleanup(interval, timeout time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			WaitingMsgsMu.Lock()
			now := time.Now()
			for msgID, msg := range WaitingMsgs {
				if now.Sub(msg.CreatedAt) > timeout {
					delete(WaitingMsgs, msgID)
				}
			}
			WaitingMsgsMu.Unlock()
		}
	}()
}
