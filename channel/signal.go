package channel

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefulExit1 优雅退出
func GracefulExit1() {
	go func() {
		fmt.Println("执行业务")
		// ...
	}()

	// 处理 "Ctrl + C" 等中断信号
	termChan := make(chan os.Signal, 1) // signal.Notify 传的 channel 要用带缓冲的 chan
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan

	// 执行推出之前的清理动作
	doCleanupQuick() // 如果 cleanup 很耗时，程序退出需要等待非常久的时间，这是不可接受的

	fmt.Println("优雅退出...")
}

func doCleanupQuick() {
	time.Sleep(1 * time.Second)
	fmt.Println("清理完毕")
}

func BetterGracefulExit() {
	closing := make(chan struct{})
	closed := make(chan struct{})

	go func() {
		// 模拟业务处理
		for {
			select {
			case <-closing:
				return
			default:
				// 业务计算
				fmt.Println("执行业务...")
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// 处理 "Ctrl + C" 等中断信号
	// the channel used with signal.Notify should be buffered (SA1017)
	termChan := make(chan os.Signal, 1) // signal.Notify 传的 channel 要用带缓冲的 chan
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan

	close(closing)
	// 执行退出前的清理操作
	go doCleanup(closed)

	select {
	case <-closed:
		fmt.Println("清理完毕")
	case <-time.After(5 * time.Second):
		fmt.Println("清理超时，不等了")
	}
	fmt.Println("优雅退出")
}

func doCleanup(closed chan struct{}) {
	fmt.Println("开始清理")
	time.Sleep(time.Minute)
	close(closed)
	fmt.Println("清理结束")
}
