package register

import (
	"StarHop/pb"
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
	Stream    pb.Stream
	CreatedAt time.Time
}

// PopWaitingMsg 取出消息
func PopWaitingMsg(msgID uint64) (pb.Stream, error) {
	WaitingMsgsMu.Lock()
	defer WaitingMsgsMu.Unlock()

	msg, exists := WaitingMsgs[msgID]
	if !exists {
		return nil, ErrMsgNotFound
	}

	delete(WaitingMsgs, msgID)
	return msg.Stream, nil
}

// 清除对应消息
func ClearWaitingMsg(msgID uint64) {
	WaitingMsgsMu.Lock()
	defer WaitingMsgsMu.Unlock()
	delete(WaitingMsgs, msgID)
}

// PutWaitingMsg 传入链接
func PutWaitingMsg(msgID uint64, stream pb.Stream) {
	WaitingMsgsMu.Lock()
	defer WaitingMsgsMu.Unlock()

	WaitingMsgs[msgID] = &waitingMsg{
		Stream:    stream,
		CreatedAt: time.Now(),
	}
}

// StartWaitingMsgsAutoCleanup 自动清理超时链接
func StartWaitingMsgsAutoCleanup(interval, timeout time.Duration) {
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
