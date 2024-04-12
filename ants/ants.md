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