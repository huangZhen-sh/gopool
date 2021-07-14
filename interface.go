package gopool

import (
	"context"
	"time"
)

type WorkerInterface interface {
	AcceptTask(task interface{}) //接收任务
	Status() bool                //标识是否已经被开除
	Tag() int                    //工人标签
	IsFired() bool               //是否会被开除
	DoFired()                    //工人被开除
}

type WorkerLeaderInterface interface {
	CWorker(b BossInterface) WorkerInterface
}

type BossInterface interface {
	BossCtx() context.Context
	AddToFreeWorkers(w WorkerInterface)
	WorkerQuantity() int
	WorkerMaxIdleTime() time.Duration
}
