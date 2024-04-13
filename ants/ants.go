package ants

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
	"sync"
	"time"
)

func myTask(i int) {
	fmt.Printf("处理任务 %d\n", i)
	time.Sleep(time.Second) // 模拟任务执行时间
}

func Demo() {
	// 使用 WaitGroup 等待所有任务完成
	var wg sync.WaitGroup

	// 创建一个协程池，池中最多有 10 个协程
	p, _ := ants.NewPoolWithFunc(10, func(i interface{}) {
		myTask(i.(int))
		wg.Done()
	})

	defer p.Release() // 程序退出时释放池

	// 添加任务到池中
	for i := 0; i < 50; i++ {
		wg.Add(1)
		_ = p.Invoke(i) // 将任务添加到池中
	}

	wg.Wait() // 等待所有任务完成
	fmt.Println("所有任务已完成")
}
