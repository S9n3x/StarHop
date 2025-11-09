package control

import (
	"StarHop/pb"
	"sync"
)

// candidateNodes 保存客户端可尝试连接的跳点节点，去重后存储。
var (
	candidateMu    sync.Mutex
	candidateCond  = sync.NewCond(&candidateMu)
	candidateIndex = make(map[string]int)
	candidateNodes []*pb.HopNodePacket
)

// StoreCandidateNodes 将新节点合并进缓存，如已存在则更新。
func StoreCandidateNodes(nodes ...*pb.HopNodePacket) {
	candidateMu.Lock()
	defer candidateMu.Unlock()

	appended := false
	for _, node := range nodes {
		if node == nil || node.Address == "" {
			continue
		}

		if idx, ok := candidateIndex[node.Address]; ok {
			candidateNodes[idx] = node
			continue
		}

		candidateIndex[node.Address] = len(candidateNodes)
		candidateNodes = append(candidateNodes, node)
		appended = true
	}

	if appended {
		candidateCond.Broadcast()
	}
}

// TakeCandidate 按先进先出取出一个可连接节点
func TakeCandidate() *pb.HopNodePacket {
	candidateMu.Lock()
	defer candidateMu.Unlock()

	for len(candidateNodes) == 0 {
		candidateCond.Wait()
	}

	node := candidateNodes[0]
	delete(candidateIndex, node.Address)

	copy(candidateNodes[0:], candidateNodes[1:])
	candidateNodes = candidateNodes[:len(candidateNodes)-1]

	for idx, n := range candidateNodes {
		candidateIndex[n.Address] = idx
	}

	return node
}

// GetAllCandidates 返回当前缓存节点的副本。
func GetAllCandidates() []*pb.HopNodePacket {
	candidateMu.Lock()
	defer candidateMu.Unlock()

	out := make([]*pb.HopNodePacket, len(candidateNodes))
	copy(out, candidateNodes)
	return out
}

// ResetCandidates 清空缓存的候选节点。
func ResetCandidates() {
	candidateMu.Lock()
	defer candidateMu.Unlock()

	candidateNodes = nil
	candidateIndex = make(map[string]int)
}
