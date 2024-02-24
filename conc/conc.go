package conc

import (
	"fmt"
	"github.com/sourcegraph/conc"
	"io"
)

// https://github.com/sourcegraph/conc
// 更优雅的并发工具，功能丰富，调用更优雅
// todo 目前尚未发布 1.0 版本，不建议在生产环境中使用，等 1.0 发布后完善此部分

func ConcDemo() {
	var wg conc.WaitGroup
	wg.Go(func() {
		// do something
		panic(io.EOF)
	})

	wg.Go(func() {
		fmt.Println("hello")
	})

	wg.WaitAndRecover()
}
