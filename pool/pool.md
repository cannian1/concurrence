# sync.Pool
## 数据结构
```go
type Pool struct {
    noCopy noCopy
    
    local     unsafe.Pointer // local fixed-size per-P pool, actual type is [P]poolLocal
    localSize uintptr        // size of the local array
    
    victim     unsafe.Pointer // local from previous cycle
    victimSize uintptr        // size of victims array
    
    // New optionally specifies a function to generate
    // a value when Get would otherwise return nil.
    // It may not be changed concurrently with calls to Get.
    New func ()
}
```

• noCopy 防拷贝标志；

• local 类型为 [P]poolLocal 的数组，数组容量 P 为 goroutine 处理器 P 的个数；

• victim 为经过一轮 gc 回收，暂存的上一轮 local；

• New 为用户指定的工厂函数，当 Pool 内存量元素不足时，会调用该函数构造新的元素.

```go
type poolLocal struct {
    poolLocalInternal
}

// Local per-P Pool appendix.
type poolLocalInternal struct {
    private any       // Can be used only by the respective P.
    shared  poolChain // Local P can pushHead/popHead; any P can popTail.
}
```
• poolLocal 为 Pool 中对应于某个 P 的缓存数据；

• poolLocalInternal.private：对应于某个 P 的私有元素，操作时无需加锁；

• poolLocalInternal.shared: 某个 P 下的共享元素链表，由于各 P 都有可能访问，因此需要加锁.

## 核心方法

### Pool.pin
```go
// pin pins the current goroutine to P, disables preemption and
// returns poolLocal pool for the P and the P's id.
// Caller must call runtime_procUnpin() when done with the pool.
func (p *Pool) pin() (*poolLocal, int) {
    // Check whether p is nil to get a panic.
    // Otherwise the nil dereference happens while the m is pinned,
    // causing a fatal error rather than a panic.
    if p == nil {
        panic("nil Pool")
    }
    
    pid := runtime_procPin()
    // In pinSlow we store to local and then to localSize, here we load in opposite order.
    // Since we've disabled preemption, GC cannot happen in between.
    // Thus here we must observe local at least as large localSize.
    // We can observe a newer/larger local, it is fine (we must observe its zero-initialized-ness).
    s := runtime_LoadAcquintptr(&p.localSize) // load-acquire
    l := p.local                              // load-consume
    if uintptr(pid) < s {
        return indexLocal(l, pid), pid
    }
    return p.pinSlow()
}
```
• pin 方法内部通过 native 方法 runtime_procPin 取出当前 P 的 index，并且将当前 goroutine 与 P 进行绑定，短暂处于不可抢占状态；

• 如果是首次调用 pin 方法，则会走进 pinSlow 方法；

• 在pinSlow 方法中，会完成 Pool.local 的初始化，并且将当前 Pool 添加到全局的 allPool 数组中，用于 gc 回收

### Pool.Get
```go
// Get selects an arbitrary item from the Pool, removes it from the
// Pool, and returns it to the caller.
// Get may choose to ignore the pool and treat it as empty.
// Callers should not assume any relation between values passed to Put and
// the values returned by Get.
//
// If Get would otherwise return nil and p.New is non-nil, Get returns
// the result of calling p.New.
func (p *Pool) Get() any {
	if race.Enabled {
		race.Disable()
	}
	l, pid := p.pin()
	// 先去获取 P 缓存数据的私有元素 private
	x := l.private
	l.private = nil // 获取到了就清空 private
	if x == nil {
		// Try to pop the head of the local shard. We prefer
		// the head over the tail for temporal locality of
		// reuse.
		x, _ = l.shared.popHead()
		if x == nil {
			x = p.getSlow(pid)
		}
	}
	runtime_procUnpin()
	if race.Enabled {
		race.Enable()
		if x != nil {
			race.Acquire(poolRaceAddr(x))
		}
	}
	if x == nil && p.New != nil {
		x = p.New()
	}
	return x
}
```
• 调用 Pool.pin 方法，绑定当前 goroutine 与 P，并且取得该 P 对应的缓存数据；

• 尝试获取 P 缓存数据的私有元素 private；

• 倘若前一步失败，则尝试取 P 缓存数据中共享元素链表的头元素；

• 倘若前一步失败，则走入 Pool.getSlow 方法，尝试取其他 P 缓存数据中共享元素链表的尾元素；

• 同样在 Pool.getSlow 方法中，倘若前一步失败，则尝试从上轮 gc 前缓存中取元素（victim）；

• 调用 native 方法解绑 当前 goroutine 与 P

• 倘若（2）-（5）步均取值失败，调用用户的工厂方法，进行元素构造并返回.

### Put
```go
// Put adds x to the pool.
func (p *Pool) Put(x any) {
	if x == nil {
		return
	}
	if race.Enabled {
		if runtime_randn(4) == 0 {
			// Randomly drop x on floor.
			return
		}
		race.ReleaseMerge(poolRaceAddr(x))
		race.Disable()
	}
	l, _ := p.pin()
	if l.private == nil {
		l.private = x
	} else {
		l.shared.pushHead(x)
	}
	runtime_procUnpin()
	if race.Enabled {
		race.Enable()
	}
}
```
• 判断存入元素 x 非空；

• 调用 Pool.pin 绑定当前 goroutine 与 P，并获取 P 的缓存数据；

• 倘若 P 缓存数据中的私有元素为空，则将 x 置为其私有元素；

• 倘若未走入（3）分支，则将 x 添加到 P 缓存数据共享链表的末尾；

• 解绑当前 goroutine 与 P.

## 回收机制

存入 pool 的对象会不定期被 go 运行时回收，因此 pool 没有容量概念，即便大量存入元素，也不会发生内存泄露.

具体回收时机是在 gc 时执行的

```go
func init() {
    runtime_registerPoolCleanup(poolCleanup)
}

func poolCleanup() {
    for _, p := range oldPools {
        p.victim = nil
        p.victimSize = 0
    }

    for _, p := range allPools {
        p.victim = p.local
        p.victimSize = p.localSize
        p.local = nil
        p.localSize = 0
    }

    oldPools, allPools = allPools, nil
}
```
• 每个 Pool 首次执行 Get 方法时，会在内部首次调用 pinSlow 方法内将该 pool 添加到迁居的 allPools 数组中；

• 每次 gc 时，会将上一轮的 oldPools 清空，并将本轮 allPools 的元素赋给 oldPools，allPools 置空；

• 新置入 oldPools 的元素统一将 local 转移到 victim，并且将 local 置为空.

综上可以得见，最多两轮 gc，pool 内的对象资源将会全被回收.