package singleflight

import (
	"context"
	"fmt"
	"golang.org/x/sync/singleflight"
	"sync"
	"time"
)

// singleflight 用于合并重复请求，减少后端压力
// Do :方法用于合并重复请求，同一个 key 的请求只会调用一次 fn 函数
// DoChan: 方法用于合并重复请求，同一个 key 的请求只会调用一次 fn 函数，但是返回一个 channel，可以和 ctx 配合使用支持超时控制
// Forget :告诉 Group 忘记 key 的请求，这样下次请求会重新调用 fn 函数

// 多个 goroutine 同时调用 Do 方法，只有一个会调用 fn 函数，其他的会等待调用结束后返回相同的结果
func Demo() {

	var wg sync.WaitGroup
	wg.Add(2)

	g := new(singleflight.Group)

	go func() {
		defer wg.Done()

		// 第三个返回参数是是否共享给其他 goroutine
		v, _, shared := g.Do("getData", func() (interface{}, error) {
			return getData(1)
		})
		fmt.Printf("one call v: %s, shared: %v\n", v.(string), shared)
	}()

	time.Sleep(time.Millisecond * 500)
	// 如果时间间隔过久等到第一个请求结束返回结果了第二个请求还没发起，
	// 源码里的 waitgroup 早就被释放了，所以第二个请求会重新调用 fn 函数，不会共享结果
	// time.Sleep(time.Second * 5)

	go func() {
		defer wg.Done()

		v, _, shared := g.Do("getData", func() (interface{}, error) {
			return getData(2)
		})
		fmt.Printf("two call v: %s, shared: %v\n", v.(string), shared)
	}()

	wg.Wait()
	// output:
	// get data
	// one call v: data_1, shared: true
	// two call v: data_1, shared: true
}

func getData(num int) (string, error) {
	fmt.Println("get data")
	time.Sleep(3 * time.Second)
	return fmt.Sprintf("data_%d", num), nil
}

// 使用 DoChan 方法，可以和 ctx 配合使用支持超时控制
// 超时的目的是为了 fail fast，快速失败以避免请求堆积
// 该示例是两个调用都被超时取消
func Demo2() {
	var wg sync.WaitGroup
	wg.Add(2)

	g := new(singleflight.Group)

	go func() {
		defer wg.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()

		v, err, shared := getDataWithTimeout(ctx, g, 1)
		fmt.Printf("one call v: %s, shared: %v err: %v\n", v, shared, err)
	}()

	time.Sleep(time.Millisecond * 500)

	go func() {
		defer wg.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()

		v, err, shared := getDataWithTimeout(ctx, g, 2)
		fmt.Printf("two call v: %s, shared: %v err: %v\n", v, shared, err)
	}()

	wg.Wait()
	// output:
	// get data
	// one call v: , shared: false err: context deadline exceeded
	// two call v: , shared: false err: context deadline exceeded
}

func getDataWithTimeout(ctx context.Context, g *singleflight.Group, num int) (string, error, bool) {
	// 用 channel 接收返回结果
	ch := g.DoChan("getData", func() (interface{}, error) {
		return getData(num)
	})

	select {
	case <-ctx.Done(): // ctx 用于超时控制
		return "", ctx.Err(), false
	case ret := <-ch:
		return ret.Val.(string), ret.Err, ret.Shared
	}
}

// 如果第一个调用超时了，后续的请求并不会共享这个超时， 而是会以当前goroutine的实际请求超时时间为准，
// 并且共享了第一个调用请求的结果，也就是只会调用下游函数一次。
// 因为这个超时是自己设置的，底层还是会调用下游函数，只是不会等待结果返回。
func Demo3() {
	g := new(singleflight.Group)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()

		v, err, shared := getDataWithTimeout(ctx, g, 1)
		fmt.Printf("one call v: %s, shared: %v err: %v\n", v, shared, err)
	}()

	time.Sleep(time.Millisecond * 2800)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()

		v, err, shared := getDataWithTimeout(ctx, g, 2)
		fmt.Printf("two call v: %s, shared: %v err: %v\n", v, shared, err)
	}()

	time.Sleep(time.Second)
	// output:
	// get data
	// one call v: , shared: false err: context deadline exceeded
	// two call v: data_1, shared: true err: <nil>
}
