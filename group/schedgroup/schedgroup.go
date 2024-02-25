package schedgroup

import (
	"context"
	"github.com/mdlayher/schedgroup"
	"log"
	"time"
)

// GopherCon Europe 2020: Matt Layher - Schedgroup: a Timer-Based Goroutine Concurrency Primitive
// https://www.youtube.com/watch?v=fWTnROKW3bg

// 少量子任务用 timer 就可以实现
// 大量子任务，且要能撤销，用 timer CPU 资源消耗就大了
// schedgroup 用 heap 结构按子任务执行时间排序，避免使用大量 timer

func SchedGroupDemo() {
	sg := schedgroup.New(context.Background())

	// 设置子任务分别在 100ms、200ms、300ms 后执行
	for i := 0; i < 3; i++ {
		n := i + 1
		sg.Delay(time.Duration(n*100)*time.Millisecond, func() {
			log.Println(n) // 输出任务编号
		})
	}

	// 等待所有子任务完成
	// 调用了 Wait 方法就不能再调 Delay 和 Schedule 方法了
	// Wait 方法只能调一次
	if err := sg.Wait(); err != nil {
		log.Fatalf("failed to wait: %v", err)
	}
}
