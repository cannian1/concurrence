package task_scheduling

import "reflect"

// 扇入模式
// 多个源 channel 输入，一个目的 channel 输出
// 可以用 反射、递归、和每一个 goroutine 处理一个 channel 的方式实现

func fanInReflect[T any](chans ...<-chan T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		var cases []reflect.SelectCase
		for _, c := range chans {
			cases = append(cases, reflect.SelectCase{
				Dir:  reflect.SelectRecv, // case 语句方向是接收
				Chan: reflect.ValueOf(c),
			})
		}

		for len(cases) > 0 {
			i, v, ok := reflect.Select(cases) // 选择一个可读取的 channel
			if !ok {                          // 如果所选择的 channel 已被关闭，则从 case 切片中剔除它
				cases = append(cases[:i], cases[i+1:]...)
				continue
			}
			out <- v.Interface().(T)
		}
	}()
	return out
}

func fanInRec[T any](chans ...<-chan T) <-chan T {
	switch len(chans) {
	case 0: // 输入 channel 的数量为 0
		c := make(chan T)
		close(c)
		return c
	case 1: // 输入 channel 的数量为 1
		return chans[0]
	case 2: // 输入 channel 的数量为 2，合并这两个 channel
		return mergeTwo(chans[0], chans[1])
	default:
		m := len(chans) / 2
		return mergeTwo(
			fanInRec(chans[:m]...), // 递归调用
			fanInRec(chans[m:]...),
		)
	}
}

func mergeTwo[T any](a, b <-chan T) <-chan T {
	c := make(chan T)

	// 使用一个 goroutine 从两个 channel 中读取数据，写入输出 channel 中
	go func() {
		defer close(c)
		for a != nil || b != nil {
			select {
			case v, ok := <-a:
				if !ok {
					a = nil
					continue
				}
				c <- v
			case v, ok := <-b:
				if !ok {
					b = nil
					continue
				}
				c <- v
			}
		}
	}()
	return c
}
