<h3>简介</h3>

go协程管理，定义最小和最大协程数，任务繁忙时多开协程数，任务空闲时多余的协程自动清除

<h3>安装</h3>

````
go get github.com/huangZhen-sh/gopool
````

<h3>使用方法</h3>

1、定义工作通道的数据类型，例如
````
type workChanType int
````

2、创建对象dw实现工人详细完成工作的接口
````
type DoWorkInterface interface {
	DetailWork(w WorkerInterface,t interface{})
}
````
3、创建boss，管理工人，监控工作通道，分配任务
```
gopool.NewBoss(10*time.Second, 20, 1, 20, 10, dw)
````
