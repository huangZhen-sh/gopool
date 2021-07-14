package gopool

import (
	"context"
	"log"
	"time"
)

type worker struct {
	tag           int       //序号标识
	startTime     time.Time //开始时间
	lastWorkTime  time.Time //最后工作时间
	ctx           context.Context
	cancel        context.CancelFunc
	status        bool
	taskChan      chan interface{} //工作通道
	bossInterface BossInterface    //向老板汇报工作的接口
	doWork        DoWorkInterface  //工人详细工作内容接口
}

func NewWorker(b BossInterface, doWork DoWorkInterface) WorkerInterface {
	pCtx := b.BossCtx()
	tag := b.WorkerQuantity()
	ctx, cancel := context.WithCancel(pCtx)
	nowTime := time.Now()
	tag++
	w := &worker{
		tag:           tag,
		startTime:     nowTime,
		lastWorkTime:  nowTime,
		ctx:           ctx,
		cancel:        cancel,
		status:        true,
		taskChan:      make(chan interface{}, 1),
		bossInterface: b,
		doWork:        doWork,
	}
	go w.listen(ctx)
	return w
}

func (w *worker) listen(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	for {
		select {
		case task, ok := <-w.taskChan:
			//具体工作
			if !ok {
				w.taskChan = nil
			} else {
				w.execute(task)
			}
		case <-ctx.Done():
			//你被解雇了
			return
			//default:
			//	fmt.Printf("工人%v==========================\n", w.tag)
			//	//time.Sleep(1 * time.Second)
		}
	}
}

func (w *worker) execute(t interface{}) {
	w.doWork.DetailWork(w, t)
	w.lastWorkTime = time.Now()
	w.bossInterface.AddToFreeWorkers(w)
}

func (w *worker) AcceptTask(t interface{}) {
	w.taskChan <- t
}

func (w *worker) Status() bool {
	return w.status
}

func (w *worker) Tag() int {
	return w.tag
}

func (w *worker) LastWorkTime() time.Time {
	return w.lastWorkTime
}

func (w *worker) DoFired() {
	w.cancel()
	w.status = false
	return
}
