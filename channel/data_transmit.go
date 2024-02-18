package channel

import (
	"fmt"
	"time"
)

// 有 4 个 goroutine，编号为 1、2、3、4。每秒有一个 goroutine 打印出自己的编号
// 下面的程序，让输出的编号总是按照 1、2、3、4、1、2、3、4……这个顺序打印出来

type Token struct{}

func newWorker(id int, ch chan Token, nextCh chan Token) {
	for {
		token := <-ch       // 获得令牌
		fmt.Println(id + 1) // id 从 1 开始
		time.Sleep(time.Second)
		nextCh <- token // 放到下一个 chan 里
	}
}

func Transmit() {
	chs := []chan Token{
		make(chan Token), make(chan Token), make(chan Token), make(chan Token),
	}

	// 创建 4 个 worker
	for i := range 4 {
		go newWorker(i, chs[i], chs[(i+1)%4])
	}

	// 首先把令牌交给第一个 worker
	chs[0] <- struct{}{}

	select {}
}
