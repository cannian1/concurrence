package sizegroup

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-pkgz/syncs"
	"sync/atomic"
	"time"
)

// 通过信号量控制并发的 goroutine 数量，或者不控制 goroutine 的数量，而是控制子任务并发执行的数量

func SizeGroupDemo() {
	// 设置 goroutine 的数量为 10
	swg := syncs.NewSizedGroup(10) // 默认处理方式
	//swg := syncs.NewSizedGroup(10, syncs.Preemptive) // 设置锁定模式，防止生成等待 goroutine 。可能导致 Go 调用阻塞
	//swg := syncs.NewSizedGroup(10, syncs.TermOnErr)  // 第一个错误发生后，不再创建新的 goroutine
	//swg := syncs.NewSizedGroup(10, syncs.Discard)    // 信号量满了之后，不再创建新的 goroutine

	var c uint32

	// 执行 1000 个子任务，同一时刻只会有 10 个 goroutine 来执行传入的函数
	for i := 0; i < 1000; i++ {
		swg.Go(func(ctx context.Context) {
			time.Sleep(5 * time.Millisecond)
			atomic.AddUint32(&c, 1)
		})
	}

	// 等待子任务完成
	swg.Wait()
	fmt.Println(c)
}

func SizeGroupErrDemo() {
	// 设置了 TermOnErr ，子任务出现第一个 error 时会撤销 Context，后面的 Go 调用会直接返回
	// Wait 调用者会得到这个错误
	swg := syncs.NewErrSizedGroup(10, syncs.TermOnErr)

	var c uint32

	for i := 0; i < 1000; i++ {
		i := i
		swg.Go(func() error {
			if i == 7 {
				return errors.New("错了")
			}
			time.Sleep(50 * time.Millisecond)
			atomic.AddUint32(&c, 1)
			return nil
		})
	}

	err := swg.Wait()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(c)
}
