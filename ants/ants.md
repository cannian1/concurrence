# Ants
## 核心数据结构
```go
// Pool accepts the tasks and process them concurrently,
// it limits the total of goroutines to a given number by recycling goroutines.
type Pool struct {
	// capacity of the pool, a negative value means that the capacity of pool is limitless, an infinite pool is used to
	// avoid potential issue of endless blocking caused by nested usage of a pool: submitting a task to pool
	// which submits a new task to the same pool.
	capacity int32 // 池子的容量

	// running is the number of the currently running goroutines.
	running int32  // 当前正在运行的 goroutine 数量

	// lock for protecting the worker queue.
	lock sync.Locker // 自制的自旋锁，保证取 goWorker 时并发安全

	// workers is a slice that store the available workers.
	workers workerQueue // goWorker 协程列表

	// state is used to notice the pool to closed itself.
	state int32 // 状态，0-打开;1-关闭

	// cond for waiting to get an idle worker.
	cond *sync.Cond // 并发协调器

	// workerCache speeds up the obtainment of a usable worker in function:retrieveWorker.
	workerCache sync.Pool // 协程对象缓存池（回收站）

	// waiting is the number of goroutines already been blocked on pool.Submit(), protected by pool.lock
	waiting int32 // 阻塞等待的协程数量

	purgeDone int32
	stopPurge context.CancelFunc

	ticktockDone int32
	stopTicktock context.CancelFunc

	now atomic.Value

	options *Options
}
```

### goWorker
```go
// goWorker is the actual executor who runs the tasks,
// it starts a goroutine that accepts tasks and
// performs function calls.
type goWorker struct {
	// pool who owns this worker.
	pool *Pool

	// task is a job should be done.
	task chan func()

	// lastUsed will be updated when putting a worker back into queue.
	lastUsed time.Time
}
```
goWorker 可以简单理解为一个长时间运行而不回收的协程，用于反复处理用户提交的异步任务，其核心字段包含：

• pool：goWorker 所属的协程池；

• task：goWorker 用于接收异步任务包的管道；

• lastUsed：goWorker 放回队列的时间.

### options
```go
// Options 包含了初始化 ants 池时将应用的所有选项。
type Options struct {
    // ExpiryDuration 是清理器协程清理过期工作协程的时间周期，
    // 清理器每隔 `ExpiryDuration` 时间扫描所有工作协程，并清理那些未被使用超过 `ExpiryDuration` 时间的协程。
    ExpiryDuration time.Duration
    
    // PreAlloc 指示在初始化池时是否进行内存预分配。
    PreAlloc bool
    
    // 最大阻塞任务数，当 pool.Submit 被调用时，如果达到这个限制，则调用者将被阻塞。
    // 0（默认值）表示没有这样的限制。
    MaxBlockingTasks int
    
    // 当 Nonblocking 为 true 时，Pool.Submit 永远不会阻塞。
    // 如果 Pool.Submit 不能立即完成，将返回 ErrPoolOverload。
    // 当 Nonblocking 为 true 时，MaxBlockingTasks 将不起作用。
    Nonblocking bool
    
    // PanicHandler 用于处理每个工作协程中的 panic。
    // 如果为 nil，panic 将再次从工作协程中抛出。
    PanicHandler func(interface{})
    
    // Logger 是用于记录信息的定制化日志记录器，如果未设置，
    // 则使用 log 包的默认标准日志记录器。
    Logger Logger
    
    // 当 DisablePurge 为 true 时，工作协程不会被清理，而是常驻。
    DisablePurge bool
}
```
协程池定制化参数集合，包含配置项如下：

• DisablePurge：是否允许回收空闲 goWorker；

• ExpiryDuration: 空闲 goWorker 回收时间间隔；仅当 DisablePurge 为 false 时有效；

• Nonblocking：是否设置为非阻塞模式，若是，goWorker 不足时不等待，直接返回 err;

• MaxBlockingTasks：阻塞模式下，最多阻塞等待的协程数量；仅当 Nonblocking 为 false 时有效；

• PanicHandler：提交任务发生 panic 时的处理逻辑；

### workerQueue
```go
type workerQueue interface {
	len() int
	isEmpty() bool
	insert(worker) error
	detach() worker
	refresh(duration time.Duration) []worker // clean up the stale workers and return them
	reset()
}
```

## 核心 API
### Pool 构造方法

```go
// NewPool instantiates a Pool with customized options.
func NewPool(size int, options ...Option) (*Pool, error) {
	if size <= 0 {
		size = -1
	}

    // 读取用户配置，做一些前置校验，默认值赋值等前处理动作
	opts := loadOptions(options...)

	if !opts.DisablePurge {
		if expiry := opts.ExpiryDuration; expiry < 0 {
			return nil, ErrInvalidPoolExpiry
		} else if expiry == 0 {
			opts.ExpiryDuration = DefaultCleanIntervalTime
		}
	}

	if opts.Logger == nil {
		opts.Logger = defaultLogger
	}

	p := &Pool{
		capacity: int32(size),
		lock:     syncx.NewSpinLock(),
		options:  opts,
	}
    // 构造 goWorker 对象池 workerCache
	p.workerCache.New = func() interface{} {
		return &goWorker{
			pool: p,
			task: make(chan func(), workerChanCap),
		}
	}
    
	// 初始化 goWorker 队列
	if p.options.PreAlloc {
		if size == -1 {
			return nil, ErrInvalidPreAllocSize
		}
		p.workers = newWorkerQueue(queueTypeLoopQueue, size)
	} else {
		p.workers = newWorkerQueue(queueTypeStack, 0)
	}

	// Pool 并发协调器
	p.cond = sync.NewCond(p.lock)

	p.goPurge()
	p.goTicktock()

	return p, nil
}
```

### Pool 提交任务

```go
// Submit submits a task to this pool.
//
// Note that you are allowed to call Pool.Submit() from the current Pool.Submit(),
// but what calls for special attention is that you will get blocked with the last
// Pool.Submit() call once the current Pool runs out of its capacity, and to avoid this,
// you should instantiate a Pool with ants.WithNonblocking(true).
func (p *Pool) Submit(task func()) error {
	if p.IsClosed() {
		return ErrPoolClosed
	}

	w, err := p.retrieveWorker()
	if w != nil {
		w.inputFunc(task)
	}
	return err
}
```

- 从 Pool 中取出一个可用的 goWorker；
- 将用户提交的任务包添加到 goWorker 的 channel 中.

```go
// retrieveWorker returns an available worker to run the tasks.
func (p *Pool) retrieveWorker() (w worker, err error) {
	p.lock.Lock()

retry:
	// 先尝试从队列中获取.
	if w = p.workers.detach(); w != nil {
		p.lock.Unlock()
		return
	}

	// worker queue 为空，且没耗尽 worker goroutine 数量
	// 接下来生成新的 goWorker.
	if capacity := p.Cap(); capacity == -1 || capacity > p.Running() {
		p.lock.Unlock()
		w = p.workerCache.Get().(*goWorker)
		w.run()
		return
	}

	// Bail out early if it's in nonblocking mode or the number of pending callers reaches the maximum limit value.
	if p.options.Nonblocking || (p.options.MaxBlockingTasks != 0 && p.Waiting() >= p.options.MaxBlockingTasks) {
		p.lock.Unlock()
		return nil, ErrPoolOverload
	}

	// Otherwise, we'll have to keep them blocked and wait for at least one worker to be put back into pool.
	p.addWaiting(1)
	p.cond.Wait() // block and wait for an available worker
	p.addWaiting(-1)

	if p.IsClosed() {
		p.lock.Unlock()
		return nil, ErrPoolClosed
	}

	goto retry
}
```



### goWorker 运行

```go
// run starts a goroutine to repeat the process
// that performs the function calls.
func (w *goWorker) run() {
	w.pool.addRunning(1)
	go func() {
		defer func() {
			w.pool.addRunning(-1)
            // 要回收之前先丢回收站
			w.pool.workerCache.Put(w)
			if p := recover(); p != nil {
				if ph := w.pool.options.PanicHandler; ph != nil {
					ph(p)
				} else {
					w.pool.options.Logger.Printf("worker exits from panic: %v\n%s\n", p, debug.Stack())
				}
			}
			// Call Signal() here in case there are goroutines waiting for available workers.
            // 回收一个 goWorker 就可以通知一个等待的 goWorke r执行
			w.pool.cond.Signal()
		}()

        
		for f := range w.task {
            // task 函数为 nil 就回收
			if f == nil {
				return
			}
			f()
            // 执行成功就放回协程池
			if ok := w.pool.revertWorker(w); !ok {
				return
			}
		}
	}()
}
```

### 回收 goWorkder

```go
// revertWorker puts a worker back into free pool, recycling the goroutines.
func (p *Pool) revertWorker(worker *goWorker) bool {
	if capacity := p.Cap(); (capacity > 0 && p.Running() > capacity) || p.IsClosed() {
		p.cond.Broadcast()
		return false
	}

	worker.lastUsed = p.nowTime()

	p.lock.Lock()
	// To avoid memory leaks, add a double check in the lock scope.
	// Issue: https://github.com/panjf2000/ants/issues/113
	if p.IsClosed() {
		p.lock.Unlock()
		return false
	}
	if err := p.workers.insert(worker); err != nil {
		p.lock.Unlock()
		return false
	}
	// Notify the invoker stuck in 'retrieveWorker()' of there is an available worker in the worker queue.
	p.cond.Signal()
	p.lock.Unlock()

	return true
}
```

### 定期销毁 goWorkder

```go
// purgeStaleWorkers clears stale workers periodically, it runs in an individual goroutine, as a scavenger.
func (p *Pool) purgeStaleWorkers(ctx context.Context) {
	ticker := time.NewTicker(p.options.ExpiryDuration)

	defer func() {
		ticker.Stop()
		atomic.StoreInt32(&p.purgeDone, 1)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		if p.IsClosed() {
			break
		}

		var isDormant bool
		p.lock.Lock()
		staleWorkers := p.workers.refresh(p.options.ExpiryDuration)
		n := p.Running()
		isDormant = n == 0 || n == len(staleWorkers)
		p.lock.Unlock()

		// Notify obsolete workers to stop.
		// This notification must be outside the p.lock, since w.task
		// may be blocking and may consume a lot of time if many workers
		// are located on non-local CPUs.
		for i := range staleWorkers {
			staleWorkers[i].finish()
			staleWorkers[i] = nil
		}

		// There might be a situation where all workers have been cleaned up (no worker is running),
		// while some invokers still are stuck in p.cond.Wait(), then we need to awake those invokers.
		if isDormant && p.Waiting() > 0 {
			p.cond.Broadcast()
		}
	}
}
```

