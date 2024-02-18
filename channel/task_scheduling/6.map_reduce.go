package task_scheduling

import "fmt"

func mapChan[T, K any](in <-chan T, fn func(T) K) <-chan K {
	out := make(chan K)
	if in == nil {
		close(out)
		return out
	}

	// 启动一个 goroutine 从输入 channel 中读取每一个数据，通过 fn 处理后发送到输出 channel 中
	go func() {
		defer close(out)

		for v := range in {
			out <- fn(v)
		}
	}()

	return out
}

func reduceChan[T any](in <-chan T, fn func(T, T) T) T {
	var out T

	if in == nil {
		return out
	}

	// 从输入 channel 中读取每一个数据，通过 fn 更新 out 的值，处理后发送到输出 channel 中
	for v := range in {
		out = fn(out, v)
	}
	return out
}

func _asStream(done <-chan struct{}) <-chan int {
	s := make(chan int)
	values := []int{1, 2, 3, 4, 5}
	go func() {
		defer close(s)

		for _, v := range values {
			select {
			case <-done:
				return
			case s <- v:
			}
		}

	}()
	return s
}

func MapReduceDemo() {
	in := _asStream(nil)

	// map 函数：每个元素乘以 10
	mapFn := func(v int) int {
		return v * 10
	}

	// reduce 函数：求和
	reduceFn := func(r, v int) int {
		return r + v
	}

	sum := reduceChan(mapChan(in, mapFn), reduceFn)
	fmt.Println(sum)
}
