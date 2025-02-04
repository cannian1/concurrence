package semaphore

import (
	"fmt"
	"sync"
	"time"
)

// 二元信号量(只有两种状态，0 和 1) 基本上就是互斥锁
// 计数信号量则是一个计数器，可以用来控制同时访问的资源的数量
// 计数信号量大于 0 时，表示有可用资源，否则表示没有可用资源

// Semaphore 数据结构，
// 书上的例子是实现了 Locker 接口，但是这仅在容量为 1时比较像，所以还是改成用信号量的常用方法
type Semaphore struct {
	// sync.Locker
	ch chan struct{}
}

// NewSemaphore 创建一个信号量
func NewSemaphore(capacity int) *Semaphore {
	if capacity <= 0 {
		capacity = 1 // 容量为 1 ，就是互斥锁
	}

	return &Semaphore{
		ch: make(chan struct{}, capacity),
	}
}

// Acquire 获取令牌
func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

// Release 释放令牌
func (s *Semaphore) Release() {
	<-s.ch
}

func customSemaphoreDemo() {
	const (
		maxConcurrent = 3  // 最大并发数
		totalTasks    = 10 // 总任务数
	)

	limiter := NewSemaphore(maxConcurrent)
	var wg sync.WaitGroup

	for i := 1; i <= totalTasks; i++ {
		// 必须等待获取到令牌才能继续
		limiter.Acquire()
		wg.Add(1)

		go func(taskID int) {
			defer limiter.Release() // 确保释放令牌
			defer wg.Done()

			fmt.Printf("[%s] Task %d started\n", time.Now().Format("15:04:05.000"), taskID)
			time.Sleep(2 * time.Second) // 模拟任务处理耗时
			fmt.Printf("[%s] Task %d done\n", time.Now().Format("15:04:05.000"), taskID)
		}(i)

		// 控制任务生成速度（可选）
		time.Sleep(200 * time.Millisecond)
	}

	wg.Wait()
}
