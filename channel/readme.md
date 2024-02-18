# channel 的应用场景

- **信息交流:** 把它当做并发的 buffer 或者队列，解决生产者-消费者问题。多个 goroutine 可以并发地作为生产者（producer）和消费者（consumer）
- **数据传递:** 一个 goroutine 将数据交给另一个 goroutine，相当于把数据的所有权交出去
- **信号通知:** 一个 goroutine 可以将信号（关闭中、已关闭、数据已准备好等）传递给了另一个或一组 goroutine
- **任务编排:** 可以让一组 goroutine 按照一定的顺序并发或者串行地执行。
- **互斥锁:** 利用 channel 也可以实现互斥锁的机制



# 行为矩阵


|  |  nil  |   非空   |    空    |    满    |   不满   |           关闭           |
|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|
| receive | 阻塞 | 读到值 | 阻塞 | 读到值 | 读到值 | 既有的值读完后，返回零值 |
| send | 阻塞 | 写入值 | 写入值 | 阻塞 | 写入值 | panic |
| close | panic | 正常关闭 | 正常关闭 | 正常关闭 | 正常关闭 | panic |

**"空读写阻塞，写关闭异常，读关闭空零"**



# 数据结构

channel 使用环形数组作为数据结构  runtime/chan.go

```go
type hchan struct {
	qcount   uint           // chan 中存在的元素数量（当前）
	dataqsiz uint           // chan 中的元素容量
	buf      unsafe.Pointer // chan中的元素队列，（环形数组）指向dataqsiz元素类型大小的数组
	elemsize uint16         // chan 中存放的每个元素的大小（声明chan 的元素类型大小）
	closed   uint32         // 标识 chan 是否关闭
	elemtype *_type         // 元素类型
	sendx    uint           // 写入元素的 index
	recvx    uint           // 读取元素的 index
	recvq    waitq          // 阻塞的读协程队列
	sendq    waitq          // 阻塞的写协程队列
	lock     mutex          // 保护hchan所有字段的锁
}
```

## waitq

```go
type waitq struct {
    first  *sudog
    last   *sudog
}
```

waitq: 阻塞的协程队列

- first：队列头部
- last：队列尾部

## sudog

```go
type sudog struct {
	// The following fields are protected by the hchan.lock of the
	// channel this sudog is blocking on. shrinkstack depends on
	// this for sudogs involved in channel ops.

	g *g

	next *sudog
	prev *sudog
	elem unsafe.Pointer // data element (may point to stack)

	// The following fields are never accessed concurrently.
	// For channels, waitlink is only accessed by g.
	// For semaphores, all fields (including the ones above)
	// are only accessed when holding a semaRoot lock.

	acquiretime int64
	releasetime int64
	ticket      uint32

	// isSelect indicates g is participating in a select, so
	// g.selectDone must be CAS'd to win the wake-up race.
	isSelect bool  // channel 在 select 内会被标识，不会陷入阻塞

	// success indicates whether communication over channel c
	// succeeded. It is true if the goroutine was awoken because a
	// value was delivered over channel c, and false if awoken
	// because c was closed.
	success bool

	parent   *sudog // semaRoot binary tree
	waitlink *sudog // g.waiting list or semaRoot
	waittail *sudog // semaRoot
	c        *hchan // channel
}
```

sudog:  用于包装协程的节点

- g：goroutine
- next：队列中的下一个节点
- prev：队列中的前一个节点
- elem：读取/写入 channel 的数据的容器
- isSelect：标识当前协程是否处在 select 多路复用的流程中
- c: 标识与当前 sudog 交互的 chan

# 构造器

```go
func makechan(t *chantype, size int) *hchan {
	elem := t.Elem

	// compiler checks this but be safe.
	if elem.Size_ >= 1<<16 {
		throw("makechan: invalid channel element type")
	}
	if hchanSize%maxAlign != 0 || elem.Align_ > maxAlign {
		throw("makechan: bad alignment")
	}

	mem, overflow := math.MulUintptr(elem.Size_, uintptr(size))
	if overflow || mem > maxAlloc-hchanSize || size < 0 {
		panic(plainError("makechan: size out of range"))
	}

	// Hchan does not contain pointers interesting for GC when elements stored in buf do not contain pointers.
	// buf points into the same allocation, elemtype is persistent.
	// SudoG's are referenced from their owning thread so they can't be collected.
	// TODO(dvyukov,rlh): Rethink when collector can move allocated objects.
	var c *hchan
	switch {
	case mem == 0:
		// Queue or element size is zero.
		c = (*hchan)(mallocgc(hchanSize, nil, true))
		// Race detector uses this location for synchronization.
		c.buf = c.raceaddr()
	case elem.PtrBytes == 0:
		// Elements do not contain pointers.
		// Allocate hchan and buf in one call.
		c = (*hchan)(mallocgc(hchanSize+mem, nil, true))
		c.buf = add(unsafe.Pointer(c), hchanSize)
	default:
		// Elements contain pointers.
		c = new(hchan)
		c.buf = mallocgc(mem, elem, true)
	}

	c.elemsize = uint16(elem.Size_)
	c.elemtype = elem
	c.dataqsiz = uint(size)
	lockInit(&c.lock, lockRankHchan)

	if debugChan {
		print("makechan: chan=", c, "; elemsize=", elem.Size_, "; dataqsiz=", size, "\n")
	}
	return c
}
```

- 保证 chan 中的元素大小小于 65535，保证内存对齐。
- 判断申请内存空间大小是否越界，mem 大小为 element 类型大小与 element 个数相乘后得到，仅当无缓冲型 channel 时，因个数为 0 导致大小为 0；
- 根据类型，初始 channel，分为 无缓冲型、有缓冲元素为 struct 型、有缓冲元素为 pointer 型 channel;
- 倘若为无缓冲型，则仅申请一个大小为默认值 96 的空间；
- 如若有缓冲的 struct 型，则一次性分配好 96 + mem 大小的空间，并且调整 chan 的 buf 指向 mem 的起始位置；
- 倘若为有缓冲的 pointer 型，则分别申请 chan 和 buf 的空间，两者无需连续；
- 对 channel 的其余字段进行初始化，包括元素类型大小、元素类型、容量以及锁的初始化.