package task_scheduling

import (
	"fmt"
	"reflect"
	"time"
)

// Or-Done 是一种信号通知模式
// "信号通知" 实现某个任务执行完成后的通知机制。在实现时，我们为这个任务定义一个类型为 chan struct{} 的 done 变量，当
// 任务执行结束后，就可以关闭这个变量，其他 receiver 就会收到这个信号，知晓任务已完成

// 例如：将同一个请求发送到多个微服务节点，只要任意一个微服务节点返回结果，就算成功。

func or(channels ...<-chan any) <-chan any {
	// 特殊情况，只有 0 个或 1 个 chan
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}

	orDone := make(chan any)
	go func() {
		defer close(orDone)

		switch len(channels) {
		case 2: // 有两个 chan，也是一种特殊情况
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			m := len(channels)
			select { // 超过两个，二分法递归处理
			case <-or(channels[:m]...):
			case <-or(channels[m:]...):
			}
		}
	}()

	return orDone
}

// 生成一个定时关闭的 channel
func sig(after time.Duration) <-chan any {
	c := make(chan any)
	go func() {
		defer close(c)
		time.Sleep(after)
	}()
	return c
}

func Demo() {
	start := time.Now()
	// 生成一组不同时间关闭的 channel，只要有一个 channel 关闭了，就往下执行
	<-or(
		sig(10*time.Second),
		sig(20*time.Second),
		sig(30*time.Second),
		sig(40*time.Second),
		sig(01*time.Second),
	)

	fmt.Println("done after", time.Since(start))
}

// 反射方式实现 Or-Done 模式，避免了深层递归
func orWithReflect(channels ...<-chan any) <-chan any {
	// 特殊情况，只有 0 个或 1 个 chan
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]

	}

	orDone := make(chan any)
	go func() {
		defer close(orDone)
		// 利用反射构建 SelectCase
		var cases []reflect.SelectCase
		for _, c := range channels {
			cases = append(cases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(c),
			})
		}

		// 选择一个可用的 case
		reflect.Select(cases)
	}()

	return orDone
}
