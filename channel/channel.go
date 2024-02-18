package channel

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

func BasicUsage() {
	ch := make(chan int, 5)
	defer close(ch)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := range 2 {
			time.Sleep(time.Duration(1+rand.Int64N(5)) * time.Second)
			ch <- i + int(rand.Int64N(7))
		}
	}()

	go func() {
		defer wg.Done()
		for i := range 21 { // 0 1 2 3 0 5 6 7 0 9 10 11 0 13 14 15 0 17 18 19 0
			if i%4 == 0 {
				ch <- 0
				continue
			}
			ch <- i
		}
	}()

	go func() {
		defer wg.Done()
		for {
			if data, ok := <-ch; ok {
				fmt.Println(data, ok)
			} else {
				fmt.Println("读关闭空零，第二个参数为 false。")
				fmt.Println(data, ok) // 0 false
				break
			}
		}
	}()

	//// 另一种遍历方式
	//for data := range ch {
	//	fmt.Print(data, " ")
	//}

	// 某个版本更新后，select 会判断当前 goroutine 是否还有可能被唤醒，如无唤醒可能则会 panic
	// select {}
	wg.Wait()
}
