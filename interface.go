package gopool

import (
	"context"
	"time"
)

// DoWorkInterface 工人具体工作内容接口
type DoWorkInterface interface {
	DetailWork(w WorkerInterface, t interface{})
}

type WorkerInterface interface {
	AcceptTask(task interface{}) //接收任务
	Status() bool                //标识是否已经被开除
	Tag() int                    //工人标签
	DoFired()                    //工人被开除
	LastWorkTime() time.Time     //最后工作时间
}

type BossInterface interface {
	BossCtx() context.Context
	AddToFreeWorkers(w WorkerInterface)
	WorkerQuantity() int
}
