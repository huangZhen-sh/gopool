##简介
go协程管理，定义最小和最大协程数，任务繁忙时多开协程数，任务空闲时多余的协程自动清除

##安装
go get github.com:huangZhen-sh/gopool

##使用方法
1、创建工人对象，需要实现如下接口
````
type WorkerInterface interface {
	AcceptTask(task interface{}) //接收任务
	Status() bool                //标识是否已经被开除
	Tag() int                    //工人标签
	IsFired() bool               //是否会被开除
	DoFired()                    //工人被开除
}
````
2、创建工头对象lw，需要实现如下接口
````
type WorkerLeaderInterface interface {
	CWorker(b BossInterface) WorkerInterface
}
````
3、创建boos，管理工人，监控工作通道，分配任务
```
gopool.NewBoss(10*time.Second, 20, 1, 20, lw, 10)
````
