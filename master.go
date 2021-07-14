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
	workerLeader        WorkerLeaderInterface
}

func NewBoss(fireTime time.Duration, maxWorkerQuantity int, minWorkerQuantity int, taskBufferSize int, wLeader WorkerLeaderInterface, chkTime int) *Boss {
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
		workerLeader:        wLeader,
		chkIdleTimeInterval: chkTime,
	}
	go w.listen(ctx)
	go w.fireWorker(ctx)
	return w
}

// Accept 接收工作任务
func (b *Boss) Accept(t interface{}) {
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
			//分配工作
			if !ok {
				task = nil
			} else {
				b.dispatchTask(task)
			}
		case <-ctx.Done():
			//下班回家
			return
		}
	}
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
				w := b.workerLeader.CWorker(b)
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
		if w.IsFired() && fireQuantity < maxFireQuantity {
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
