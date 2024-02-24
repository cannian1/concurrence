package waitgroup

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// waitGroup 是一个计数信号量，用来等待一组 goroutine 完成
// 一个 WaitGroup 的零值没有用，必须使用 Add 方法设置计数值
// Add 方法设置计数值为正数，Done 方法将计数值减 1，Wait 方法会阻塞直到计数值为 0
// Add 方法设置计数值为负数会引发 panic
// Done 方法将计数值减 1，如果计数值为 0 会引发 panic
// Wait 方法会阻塞直到计数值为 0
// WaitGroup 没有能力控制执行任务的 goroutine 终止，只能等待它们完成

func WaitGroupDemo() {
	var wg sync.WaitGroup
	// 对于一个零值的 WaitGroup，或者计数值已经为 0 的 WaitGroup，如果直接调用它的 Wait 方法，不会阻塞
	wg.Wait()
	wg.Wait() // 多次调用也不会有问题

	var urls = []string{"http://baidu.com", "http://bing.com", "http://google.com"}
	var result = make([]bool, len(urls))
	http.DefaultClient.Timeout = time.Second

	wg.Add(3) // 设置计数值为 3
	for i := 0; i < 3; i++ {
		i := i
		go func(url string) {
			defer wg.Done()

			log.Println("fetching ", url)
			resp, err := http.Get(url)
			if err != nil {
				result[i] = false
				return
			}
			result[i] = resp.StatusCode == http.StatusOK
			resp.Body.Close()
		}(urls[i])
	}

	wg.Wait()
	log.Println("done") // 子任务完成，result 中保证有值

	for i := 0; i < 3; i++ {
		log.Println(urls[i], ":", result[i])
	}
}
