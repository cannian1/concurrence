package errgroup

import (
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"time"
)

// ReturnFirstErrDemo 返回第一个错误
func ReturnFirstErrDemo() {
	// 启动三个任务，只有三个任务都执行完成后，才会返回第二个任务的错误

	// WithContext 返回一个带有取消信号的上下文
	// eg, ctx := errgroup.WithContext(context.Background())

	var eg errgroup.Group
	// 启动第一个子任务，它执行成功
	eg.Go(func() error {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("exec #1")
		return nil
	})

	// 执行第二个子任务，它执行失败
	eg.Go(func() error {
		time.Sleep(1000 * time.Millisecond)
		fmt.Println("exec #2")
		return errors.New("fail to exec #2")
	})

	// 执行第三个子任务，它执行成功
	eg.Go(func() error {
		time.Sleep(1500 * time.Millisecond)
		fmt.Println("exec #3")
		return nil
	})

	// 等待所有任务执行完成
	if err := eg.Wait(); err != nil {
		fmt.Println("failed:", err)
		return
	} else {
		fmt.Println("succeed")
	}
}

// ReturnAllErrDemo 返回所有错误
func ReturnAllErrDemo() {
	// Group 只能返回第一个错误，如果需要返回所有错误，可以使用切片来收集错误
	var eg errgroup.Group
	result := make([]error, 3)

	// 启动第一个子任务，它执行成功
	eg.Go(func() error {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("exec #1")
		result[0] = nil // 保存执行成功或失败的结果
		return nil
	})

	// 执行第二个子任务，它执行失败
	eg.Go(func() error {
		time.Sleep(1000 * time.Millisecond)
		fmt.Println("exec #2")
		result[1] = errors.New("fail to exec #2")
		return result[1]
	})

	// 执行第三个子任务，它执行成功
	eg.Go(func() error {
		time.Sleep(1500 * time.Millisecond)
		fmt.Println("exec #3")
		result[2] = nil
		return nil
	})

	if err := eg.Wait(); err != nil {
		fmt.Printf("failed:%v\n", result)
	} else {
		fmt.Println("succeed")
	}
}

// todo 使用 rabbitmq 的例子
func SetLimitDemo() {

}
