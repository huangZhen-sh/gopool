package gopool

import (
	"context"
	"time"
)

// BossCtx 招聘工人时就告知对方，如果老板破产、工人自然也要跟着被解雇
func (b *Boss) BossCtx() context.Context {
	return b.ctx
}

func (b *Boss) WorkerQuantity() int {
	return len(b.workers)
}

// AddToFreeWorkers 工人向老板汇报工作的接口
func (b *Boss) AddToFreeWorkers(w WorkerInterface) {
	b.freeWorkers <- w
}


func (b *Boss) WorkerMaxIdleTime()time.Duration {
	return b.workerMaxIdleTime
}
