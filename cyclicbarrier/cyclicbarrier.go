package cyclicbarrier

import (
	"context"
	"fmt"
	"github.com/marusama/cyclicbarrier"
	"log"
	"math/rand/v2"
	"sync"
	"time"
)

// 同步屏障(Barrier)是一种同步机制。
// 对于一组 goroutine 程序中的一个或多个 goroutine 被阻塞在屏障处，直到所有 goroutine 都到达屏障位置。

// CyclicBarrier 允许一组 goroutine 互相等待，直到到达某个公共检查点，然后到达下一个同步点，循环使用。
// 因为屏障在释放等待的 goroutine 之后可以重用，所以它被称为循环屏障。

// CyclicBarrier VS  WaitGroup
// ----------------------------
// New(n)            var wg WaitGroup
//                   wg.Add(n)
// ----------------------------
// Await()           wg.Done()
//                   wg.Wait()
//                   wg.Add(n) // 重用
// ----------------------------

// WaitGroup 适合用在一个 goroutine 等待一组 goroutine 到达同一个检查点
// CyclicBarrier 的参与者互相等待，WaitGroup一般是父 goroutine 等待子 goroutine 完成，子 goroutine 之间不需要相互等待

func Demo() {
	cnt := 0
	b := cyclicbarrier.NewWithAction(10, func() error {
		cnt++
		return nil
	})

	wg := sync.WaitGroup{}
	wg.Add(10)

	for i := 0; i < 10; i++ { // 启动 10 个 goroutine
		i := i
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ { // 执行 5 轮
				time.Sleep(time.Duration(rand.IntN(10)) * time.Second)
				// 每一轮随机休眠一段时间，再来到屏障处
				log.Printf("goroutine %d, 来到第 %d 轮屏障, wait\n", i, j)
				err := b.Await(context.TODO())
				log.Printf("goroutine %d, 冲破第 %d 轮屏障\n", i, j)
				if err != nil {
					panic(err)
				}
			}
		}()
	}

	wg.Wait()
	fmt.Println(cnt)
}
