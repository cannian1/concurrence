package task_scheduling

import "golang.org/x/exp/constraints"

// 把流穿起来，就产生了管道模式

func sqrt[T constraints.Integer](in <-chan T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for v := range in {
			out <- v * v
		}
	}()
	return out
}
