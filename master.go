package gopool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Boss struct {
	workerMaxIdleTime   time.Duration        //工人最大空闲时间，超过这个时间将被开除
	chkIdleTimeInterval int                  //定期评估工人是否有效单位秒
	maxWorkerNum        int                  //最大工人数量
	minWorkerNum        int                  //最少工人数量
	taskChan            chan interface{}     //工作通道
	workers             []WorkerInterface    //工人切片集合
	freeWorkers         chan WorkerInterface //空闲工人数
	lock                sync.Mutex           //互斥锁
	ctx                 context.Context
	cancel              context.CancelFunc
	doWork              DoWorkInterface //工人详细工作内容接口
	waitStopChan        chan bool       //等待结束
}

func NewBoss(fireTime time.Duration, maxWorkerQuantity int, minWorkerQuantity int, taskBufferSize int, chkTime int, doWork DoWorkInterface) *Boss {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Boss{
		workerMaxIdleTime:   fireTime,
		maxWorkerNum:        maxWorkerQuantity,
		minWorkerNum:        minWorkerQuantity,
		taskChan:            make(chan interface{}, taskBufferSize),
		workers:             make([]WorkerInterface, 0, maxWorkerQuantity),
		freeWorkers:         make(chan WorkerInterface, maxWorkerQuantity),
		ctx:                 ctx,
		cancel:              cancel,
		chkIdleTimeInterval: chkTime,
		doWork:              doWork,
		waitStopChan:        make(chan bool),
	}
	go w.listen(ctx)
	go w.fireWorker(ctx)
	return w
}

// Accept 接收工作任务
func (b *Boss) Accept(t interface{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	b.taskChan <- t
}

//监听工作任务
func (b *Boss) listen(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	for {
		select {
		case task, ok := <-b.taskChan:
			if !ok {
				//通道关闭，所有协程全部退出
				b.taskChan = nil
				b.cancel()
				b.waitStopChan <- true
			} else {
				//分配工作
				b.dispatchTask(task)
			}
		case <-ctx.Done():
			//下班回家
			return
		}
	}
}

// Stop 关闭任务通道，等待，剩余任务完成后，所有协程全部退出
func (b *Boss) Stop() bool {
	close(b.taskChan)
	<-b.waitStopChan
	return true
}

//给工人分配工作
func (b *Boss) dispatchTask(task interface{}) {
	w := b.callWorker()
	fmt.Printf("%v工人接收工作\n", w.Tag())
	w.AcceptTask(task)
}

//呼叫工人
func (b *Boss) callWorker() WorkerInterface {
	for {
		select {
		case w, ok := <-b.freeWorkers:
			if !ok {
				//通道关闭置成nil,不会再进入此case
				b.freeWorkers = nil
			} else {
				if w.Status() {
					return w
				}
			}
		default:
			//如果没有已雇佣的工人，重新雇佣一个
			wLen := len(b.workers)
			if b.maxWorkerNum > wLen {
				w := NewWorker(b, b.doWork)
				b.lock.Lock()
				b.workers = append(b.workers, w)
				b.lock.Unlock()
				b.freeWorkers <- w
			}
		}
	}
}

//毎10分钟进行一次淘汰算法，开除部分工人
func (b *Boss) fireWorker(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	for {
		select {
		case <-time.After(time.Duration(b.chkIdleTimeInterval) * time.Second):
			b.doFireWorker()
		case <-ctx.Done():
			return
		}
	}
}

//开除多余的工人
func (b *Boss) doFireWorker() {
	if len(b.workers) <= b.minWorkerNum {
		return
	}
	maxFireQuantity := len(b.workers) - b.minWorkerNum
	fireQuantity := 0
	for _, w := range b.workers {
		if time.Now().Sub(w.LastWorkTime()) > b.workerMaxIdleTime && fireQuantity < maxFireQuantity {
			fireQuantity++
			w.DoFired()
			fmt.Printf("%v工人被开除\n", w.Tag())
		}
	}
	if fireQuantity > 0 {
		b.lock.Lock()
		activeWorkers := make([]WorkerInterface, 0)
		for _, w := range b.workers {
			if w.Status() {
				activeWorkers = append(activeWorkers, w)
			}
		}
		b.workers = activeWorkers
		b.lock.Unlock()
	}
}

// BossCtx 招聘工人时就告知对方，如果老板破产或者不想干了、工人自然也要跟着被解雇
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
