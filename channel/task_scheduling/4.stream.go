package task_scheduling

// 把 channel 当做流式管道使用

// 把一组数据 values 转换成一个 channel
func asStream[T any](done <-chan struct{}, values ...T) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		for _, v := range values {
			select {
			case <-done:
				return
			case out <- v:
			}
		}
	}()
	return out
}

// 只取流中的前 n 个数据
func takeN[T any](done <-chan struct{}, in <-chan T, num int) <-chan T {
	takeStream := make(chan T)

	go func() {
		defer close(takeStream)

		for i := 0; i < num; i++ {
			select {
			case <-done:
				return
			case v := <-in:
				takeStream <- v
				//case takeStream <- <-in: // 也可以这样写，不过这样写会导致无法处理 done 信号
			}
		}
	}()

	return takeStream
}
