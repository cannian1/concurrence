package task_scheduling

// 扇出模式
// 一个源 channel 输出，多个目的 channel 输入
// 扇出模式经常用在设计模式的观察者模式中

// 从源 channel 中读取数据，依次发送给多个目的 channel——可以同步或异步发送
func fanOut[T any](ch <-chan T, out []chan T, async bool) {
	go func() {
		defer func() {
			for _, c := range out {
				close(c)
			}
		}()
	}()

	for v := range ch { // 从输入 channel 中读取数据, 发送给各个 channel 中
		v := v // go 1.22 后不需要这么做了
		for _, c := range out {
			c := c
			if async {
				go func() { c <- v }()
			} else {
				c <- v
			}
		}
	}
}
