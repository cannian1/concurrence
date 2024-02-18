package context

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// 父 ctx 控制子 ctx
func TestParentCancelCtx(t *testing.T) {
	ctx := context.Background()
	deadlineCtx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	childCtx := context.WithValue(deadlineCtx, "key", 123)
	cancel()
	// 上层 ctx cancel 之后，下层也会 cancel
	err := childCtx.Err()
	fmt.Println(err)
}

// 子 ctx 可以得到父 ctx 的值，反之不可以
func TestParentValueCtx(t *testing.T) {
	ctx := context.Background()
	childCtx := context.WithValue(ctx, "key1", 123)
	ccCtx := context.WithValue(childCtx, "key2", 124)

	// 上层拿不到下层的值
	fmt.Println("上层访问下层的值", childCtx.Value("key2"))
	// 下层可以拿到上层的值
	fmt.Println("下层访问上层的值", ccCtx.Value("key1"))

	// 有很吊诡的做法可以破坏 context 的不可变性
	// 把 map 或者结构体指针当做 value 塞到上级 context
	// 下级 context 通过 key 取出这个 map
	// 然后用类型断言之后赋值
	// 这样上层也可以访问到变化后的 context 的 value
	// 只有不得已父 context 不得不获取子 context 的值的时候才这么做
	lv1Ctx := context.WithValue(ctx, "map", map[string]int{})
	lv2Ctx := context.WithValue(lv1Ctx, "key3", "value3")
	m := lv2Ctx.Value("map").(map[string]int)
	m["key3"] = 114514
	fmt.Println("正常父ctx访问子ctx", lv1Ctx.Value("key3"))
	fmt.Println("破坏不可变性后", lv1Ctx.Value("map"))
}

func TestContestTimeOut(t *testing.T) {
	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	time.Sleep(2 * time.Second)
	// 判断什么类型导致 ctx的错误
	err := timeoutCtx.Err()
	switch {
	case errors.Is(err, context.Canceled):
		fmt.Println("context 被取消, err:", err)
	case errors.Is(err, context.DeadlineExceeded):
		fmt.Println("context 超时, err:", err)
	}
}

// 子 ctx 控制重置超时时间无效
func TestContestTimeOut2(t *testing.T) {
	ctx := context.Background()
	timeoutCtx, cancel1 := context.WithTimeout(ctx, time.Second)
	// 子 ctx试图重新设置超时时间，然而没有成功
	subCtx, cancel2 := context.WithTimeout(timeoutCtx, 3*time.Second)
	go func() {
		// 一秒后就会过期，输出timeout
		<-subCtx.Done()
		fmt.Println("time out")
	}()

	time.Sleep(2 * time.Second)
	cancel2()
	cancel1()
}

// 控制业务超时
func TestBusinessTimeout(t *testing.T) {
	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	end := make(chan struct{}, 1)
	go func() {
		MyBusiness()
		end <- struct{}{}
	}()
	ch := timeoutCtx.Done()
	select {
	case <-ch:
		fmt.Println("超时")
	case <-end:
		fmt.Println("业务正常结束")
	}
	// time.AfterFunc一般用于定时任务，而不是超时控制
	// 如果不主动取消，AfterFunc必然执行
	// 即使 timer.Stop() 主动取消，会有一小段时间差
}

func MyBusiness() {
	time.Sleep(1 * time.Second)
	// time.Sleep(500 * time.Millisecond)
	fmt.Println("你好世界")
}

func TestContestCancel(t *testing.T) {
	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second)
	time.Sleep(500 * time.Millisecond)
	cancel()
	err := timeoutCtx.Err()
	switch {
	case errors.Is(err, context.Canceled):
		fmt.Println("context被取消", err)
	case errors.Is(err, context.DeadlineExceeded):
		fmt.Println("context超时", err)
	}
}

// 超时时间
func TestContestDeadline(t *testing.T) {
	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	// Deadline 返回应取消此上下文的时间。
	// 当没有设置截止日期时，截止日期返回 ok==false。连续调用 Deadline 返回相同的结果。
	deadline, ok := timeoutCtx.Deadline()
	fmt.Println(deadline, ok) // 超时时间，是否设置了超时时间
}

func TestContestValue(t *testing.T) {
	ctx := context.Background()
	valCtx := context.WithValue(ctx, "abc", 123)
	val := valCtx.Value("abc")
	fmt.Println(val) // 123
}

// 带取消原因
func TestWithCause(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	// cancel(nil) // 原因如果是 nil，则和普通的 WithCancel 方法相同
	cancel(errors.New("怎么忍心怪你犯了错，是我对你做了过了火"))

	fmt.Println(ctx.Err())
	fmt.Println(context.Cause(ctx))
}

func TestAfterFunc(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	context.AfterFunc(ctx, func() {
		fmt.Println("执行 context 结束后的回调函数...")
	})
	cancel()
}
