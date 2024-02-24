package once

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// 还可以封装在结构体里，参考 math/big/sqrt.go 的 threeOnce

func Demo() {
	var o sync.Once

	// 第一个初始化函数
	f1 := func() {
		fmt.Println("f1")
	}
	o.Do(f1)

	// 第二个初始化函数
	f2 := func() {
		fmt.Println("f2")
	}
	o.Do(f2) // 无输出

	// 如果 执行的函数 panic，每次调用都会 panic 返回相同的值
	f3 := sync.OnceFunc(func() {
		fmt.Println("f3 执行")
		// panic("虾米诺手")
	})
	f3()
	f3() // 可以并发调用，再次调用不会被执行

	// OnceValue 返回函数的执行结果
	f4 := sync.OnceValue(func() string {
		return time.Now().Format("2006-01-02 15:04:05.000")
	})

	r1 := f4()
	r2 := f4()
	fmt.Println("r1 is ", r1, "|r2 is ", r2)
	// r1 和 r2 获取到的值是相同的，即 f4只执行了一次

	var jsonBlob = []byte(`[
	{"Name": "Platypus", "Order": "Monotremata"},
	{"Name": "Quoll",    "Order": "Dasyuromorphia"}
]`)
	type Animal struct {
		Name  string
		Order string
	}
	// OnceValues 返回俩泛型变量，可以用来返回 函数执行结果 和 error
	f5 := sync.OnceValues(func() ([]Animal, error) {
		var animals []Animal
		err := json.Unmarshal(jsonBlob, &animals)
		if err != nil {
			return nil, err
		}
		return animals, nil
	})

	r3, e1 := f5()
	fmt.Println(r3, e1)
}
