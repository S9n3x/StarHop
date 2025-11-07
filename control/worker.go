package control

var workerPool *WorkerPool

type tunnelMsg struct {
	id   uint64
	data []byte
}

type WorkerPool struct {
	taskChan chan tunnelMsg
	workers  int
}

func initWorkerPool(workers int, queueSize int) {
	workerPool = &WorkerPool{
		taskChan: make(chan tunnelMsg, queueSize),
		workers:  workers,
	}

	for i := 0; i < workers; i++ {
		go workerPool.worker()
	}
}

func (p *WorkerPool) worker() {
	for msg := range p.taskChan {
		receiveTunnelData(msg)
	}
}

func (p *WorkerPool) Submit(data tunnelMsg) {
	p.taskChan <- data
}

func SubmitPackage(id uint64, data []byte) {
	workerPool.Submit(tunnelMsg{
		id:   id,
		data: data,
	})
}
