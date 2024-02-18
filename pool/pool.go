package pool

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

func Demo() {
	// 创建一个对象池
	var p sync.Pool
	// 对象创建的方法
	p.New = func() any {
		return &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	var wg sync.WaitGroup
	// 使用 10 个 goroutine 测试从对象池中获取和放回对象
	wg.Add(10)
	go func() {
		for i := range 10 {
			go func() {
				defer wg.Done()
				// 获取一个对象，如果不存在则创建一个
				c := p.Get().(*http.Client)
				// 使用完毕后放回池子
				defer p.Put(c)

				resp, err := c.Get("https://bing.com")
				if err != nil {
					fmt.Println("请求 bing 失败, err:", err)
					return
				}

				resp.Body.Close()
				fmt.Println("请求成功", i) // go 1.22 for 循环每次创建新的循环变量
			}()
		}
	}()

	wg.Wait()
}

func Demo2() {
	var p sync.Pool

	// 没有设置 New ，初始化时放入 5 个可重用对象，池中最多只有 5 个对象
	// 不设置 New，池化对象如果长时间没有被使用，可能会被回收，被回收后只能获得值为 nil 的结果
	for _ = range 5 {
		p.Put(&http.Client{Timeout: 5 * time.Second})
	}

	// 模拟对象被回收，可以把这两行注释掉看看区别
	// 至于为什么要 GC 两次，参考 sync/pool.go 的 poolCleanup() 方法清除两级缓存，对象将在两个 GC周期被释放
	// [深度解密 Go 语言之 sync.Pool]https://www.cnblogs.com/qcrao-2018/p/12736031.html

	// go 1.13 之后，pool 使用无锁队列，内部有 pin() 方法把当前 goroutine 固定在当前 P 上,每个 P 只有一个活动的 g 运行，因此无需加锁
	runtime.GC()
	runtime.GC()

	var wg sync.WaitGroup
	// 使用 10 个 goroutine 测试从对象池中获取和放回对象
	wg.Add(10)
	go func() {
		for i := range 10 {
			go func() {
				defer wg.Done()

				c, ok := p.Get().(*http.Client)
				if !ok { // 可能从池中获取不到对象
					fmt.Println("client 为 nil")
					return
				}
				// 使用完毕后放回池子
				defer p.Put(c)

				resp, err := c.Get("https://bing.com")
				if err != nil {
					fmt.Println("请求 bing 失败, err:", err)
					return
				}

				resp.Body.Close()
				fmt.Println("请求成功", i)
			}()
		}
	}()

	wg.Wait()

}
