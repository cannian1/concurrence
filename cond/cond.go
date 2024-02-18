package cond

import (
	"fmt"
	"log"
	"math/rand/v2"
	"sync"
	"time"
)

func Demo() {
	// Cond 通常用于等待某个条件的一组的 goroutine
	c := sync.NewCond(&sync.Mutex{})

	var ready int
	for i := range 10 {
		go func(i int) {
			time.Sleep(time.Duration(rand.Int64N(10)) * time.Second)

			// 加锁，更改等待的条件
			c.L.Lock()
			ready++
			c.L.Unlock()

			log.Printf("运动员#%d已准备就绪\n", i)
			// 唤醒 notifyList 中所有等待此 Cond 的 goroutine
			// c.Signal() // 唤醒 notifyList 的第一个 goroutine
			c.Broadcast() // 唤醒 notifyList 的所有 goroutine

		}(i)
	}

	// 循环判断没有满足条件就 wait
	c.L.Lock()
	for ready != 10 {
		c.Wait() // wait 方法需要持有锁才能调用，可以有多个 goroutine 在 wait
		log.Println("裁判员被唤醒一次")
	}
	c.L.Unlock()

	// 所有运动员是否准备就绪
	log.Println("一切就绪")

}

// -----------------------------
var sharedRsc = make(map[string]any)

func Demo2() {

	m := sync.Mutex{}
	c := sync.NewCond(&m)
	go func() {
		// this go routine wait for changes to the sharedRsc
		c.L.Lock()
		for len(sharedRsc) == 0 {
			c.Wait() // 被通知之后就到下一次 for 循环了，发现条件变化就不再进入循环
		}
		fmt.Println(sharedRsc["rsc1"])
		c.L.Unlock()
	}()

	go func() {
		// this go routine wait for changes to the sharedRsc
		c.L.Lock()
		for len(sharedRsc) == 0 {
			c.Wait()
		}
		fmt.Println(sharedRsc["rsc2"])
		c.L.Unlock()
	}()

	// 确保两个 goroutine 都进入了循环，在 c.Wait() 处阻塞，以免因为父 goroutine 改变监听条件直接执行
	time.Sleep(1 * time.Second)

	// this one writes changes to sharedRsc
	c.L.Lock()
	sharedRsc["rsc1"] = "foo"
	sharedRsc["rsc2"] = "bar"
	fmt.Println(len(sharedRsc))
	c.Broadcast()
	c.L.Unlock()

	time.Sleep(3 * time.Second)
}

// -----------------------------
var done = false

func read(name string, c *sync.Cond) {
	c.L.Lock()
	for !done {
		c.Wait()
	}
	log.Println(name, "starts reading")
	c.L.Unlock()
}

func write(name string, c *sync.Cond) {
	log.Println(name, "starts writing")
	time.Sleep(time.Second)
	c.L.Lock()
	done = true
	c.L.Unlock()
	log.Println(name, "wakes all")
	c.Broadcast()
}

func Demo3() {
	cond := sync.NewCond(&sync.Mutex{})

	go read("reader1", cond)
	go read("reader2", cond)
	go read("reader3", cond)
	write("writer", cond)

	time.Sleep(time.Second * 3)
}
