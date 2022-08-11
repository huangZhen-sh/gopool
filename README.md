<h3>简介</h3>

go协程管理，定义最小和最大协程数，任务繁忙时自动多开协程数，任务空闲时自动清除多余的协程

<h3>安装</h3>

````
go get github.com/huangZhen-sh/gopool
````

<h3>使用方法</h3>

````
//1）定义数据结构

type ChanType int

//2) 定义具体的工作内容

type DetailWork struct{}

func (dw *DetailWork) DetailWork(w gopool.WorkerInterface, t interface{}) {
	if task, ok := t.(ChanType); ok {
		log.Printf("工人%v,会任务%v...", w.Tag(), task)
		time.Sleep(10 * time.Second)
	} else {
		log.Printf("工人%v,不会任务%v...", w.Tag(), t)
	}
}

//3）创建老板

func main() {
	boss := gopool.NewBoss(10*time.Second, 10, 1, 1000, 600, &DetailWork{})
	boss.Accept(ChanType(10))
}
````