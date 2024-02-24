package semaphore

import (
	"context"
	"fmt"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/semaphore"
	"runtime"
	"time"
)

// golang.org/x/sync/semaphore

// NewWeighted 初始化包含 n 个资源的信号量
// Acquire 请求 n 个资源，会阻塞直到有足够的资源可用或 ctx 被取消，返回 nil 表示成功，返回 ctx.Err() 表示失败
// Release 释放 n 个资源
// TryAcquire 尝试请求 n 个资源，不会阻塞，返回 true 表示成功，返回 false 表示失败

// 和 channel 相比，信号量可以一次请求/释放多个资源，而 channel 只能一次请求/释放一个资源

// 如果需要动态修改信号量的容量，可以使用 https://github.com/marusama/semaphore 包

func Demo() {

	var (
		maxWorkers = runtime.GOMAXPROCS(0)                    // 最大并发数与 CPU 核心数相同
		sema       = semaphore.NewWeighted(int64(maxWorkers)) // 信号量
		task       = make([]int, maxWorkers*4)                // 任务数量，是最大并发数的 4 倍
	)

	ctx := context.Background()

	for i := range task {
		// 如果没有 worker 可用，会阻塞直到有 worker 被释放
		if err := sema.Acquire(ctx, 1); err != nil {
			break
		}

		// 启动 worker
		go func(i int) {
			defer sema.Release(1)
			// do something
			time.Sleep(100 * time.Millisecond)
			task[i] = i + 1
		}(i)
	}

	// 请求所有 worker，这样能确保所有 worker 完成
	if err := sema.Acquire(ctx, int64(maxWorkers)); err != nil {
		// handle error
		slog.Error("获取所有 worker 失败,", "err:", err)
	}
	fmt.Println(task)
}
