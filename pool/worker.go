package pool

import "sync"

type Pool[T any] struct {
	taskQueue chan T  // 任务队列
	taskFn    func(T) // 任务的执行函数
	workers   int     // worker 的数量
	wg        sync.WaitGroup
}

// NewPool 创建一个新的 worker 池
func NewPool[T any](workers, capacity int, taskFn func(T)) *Pool[T] {
	pool := &Pool[T]{
		taskQueue: make(chan T, capacity),
		taskFn:    taskFn,
		workers:   workers,
	}
	pool.wg.Add(workers)

	return pool
}

// Start 启动 worker 池
func (p *Pool[T]) Start() {
	for _ = range p.workers {
		go func() {
			defer p.wg.Done()

			for {
				task, ok := <-p.taskQueue // 从任务队列中读取一个任务
				if !ok {                  // channel 已关闭，并且任务都已经处理完了
					return
				}
				p.taskFn(task)
			}
		}()
	}
}

// Submit 提交任务
func (p *Pool[T]) Submit(task T) {
	p.taskQueue <- task
}

func (p *Pool[T]) Close() {
	close(p.taskQueue)
	p.wg.Wait()
}
