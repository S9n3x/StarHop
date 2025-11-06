package control

var workerPool *WorkerPool

type WorkerPool struct {
	taskChan chan []byte
	workers  int
}

func initWorkerPool(workers int, queueSize int) {
	workerPool = &WorkerPool{
		taskChan: make(chan []byte, queueSize),
		workers:  workers,
	}

	for i := 0; i < workers; i++ {
		go workerPool.worker()
	}
}

func (p *WorkerPool) worker() {
	for data := range p.taskChan {
		receiveTunnelData(data)
	}
}

func (p *WorkerPool) Submit(data []byte) {
	p.taskChan <- data
}

func SubmitPackage(data []byte) {
	workerPool.Submit(data)
}
