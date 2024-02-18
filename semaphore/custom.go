package semaphore

import "sync"

// 二元信号量(只有两种状态，0 和 1) 基本上就是互斥锁
// 计数信号量则是一个计数器，可以用来控制同时访问的资源的数量
// 技术信号量大于 0 时，表示有可用资源，否则表示没有可用资源

// Semaphore 数据结构，还实现了 Locker 接口
type Semaphore struct {
	sync.Locker
	ch chan struct{}
}

// NewSemaphore 创建一个信号量
func NewSemaphore(capacity int) sync.Locker {
	if capacity <= 0 {
		capacity = 1 // 容量为 1 ，就是互斥锁
	}

	return &Semaphore{
		ch: make(chan struct{}, capacity),
	}
}

// Lock 请求一个资源
func (s *Semaphore) Lock() {
	s.ch <- struct{}{}
}

// Unlock 释放一个资源
func (s *Semaphore) Unlock() {
	<-s.ch
}
