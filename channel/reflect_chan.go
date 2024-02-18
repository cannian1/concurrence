package channel

import (
	"fmt"
	"reflect"
)

// ReflectChanDemo 利用反射在 select 里动态创建 case 处理 channel
func ReflectChanDemo() {
	var ch1 = make(chan int, 10)
	var ch2 = make(chan int, 10)

	// 创建 SelectCase
	var cases = createCases(ch1, ch2)

	// 执行 10 次 select，从 cases 中选一个 case 执行。第一次选择肯定是 send case，因为此时 chan 无元素，recv 不可用
	for range 10 {
		// reflect.Select 可以传入一组运行时的 case 语句，当做参数执行 （最多 65535 个 case）
		// chosen 是 被选中 case 的索引。如果没有可用的 case，则会返回一个 bool 类型的值，用于表示是否有 case 被成功选择
		chosen, recv, ok := reflect.Select(cases)
		if recv.IsValid() { // recv case
			fmt.Println("recv:", cases[chosen].Dir, recv, ok)
		} else { // send case
			fmt.Println("send:", cases[chosen].Dir, ok)
		}
	}
}

// 利用反射创建 case ，分别为每个 chan 生成了 recv case 和 send case，并返回 []reflect.SelectCase
func createCases(chs ...chan int) []reflect.SelectCase {
	var cases []reflect.SelectCase

	// 创建 recv case
	for _, ch := range chs {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		})
	}

	// 创建 send case
	for i, ch := range chs {
		v := reflect.ValueOf(i)
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(ch),
			Send: v,
		})
	}
	return cases
}
