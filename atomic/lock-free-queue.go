package atomic

import (
	"sync/atomic"
	"unsafe"
)

// LKQueue 是以 lock-free 方式实现的队列，它只需要 head 和 tail 两个字段
type LKQueue[T any] struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

// 队列中的每个节点，除自己的值以外，还有 next 字段指向下一个节点
type node[T any] struct {
	value T
	next  unsafe.Pointer
}

func NewLKQueue[T any]() *LKQueue[T] {
	n := unsafe.Pointer(&node[T]{})
	return &LKQueue[T]{
		head: n,
		tail: n,
	}
}

// Enqueue 表示入队
func (q *LKQueue[T]) Enqueue(v T) {
	n := &node[T]{value: v}

	// 入队通过 cas 操作将一个元素添加到队尾，并且移动 tail（尾）指针
	for {
		tail := load[T](&q.tail)
		next := load[T](&tail.next)

		if tail == load[T](&q.tail) { // tail 和 next 是否一致
			if next == nil {
				if cas(&tail.next, next, n) {
					cas(&q.tail, tail, n) // 入队完成，设置 tail
					return
				}
			} else {
				cas(&q.tail, tail, next)
			}
		}
	}
}

// Dequeue 表示出队
// 出队的时候，移除一个节点，并通过 CAS 操作移动 head 指针，同时在必要的时候移动 tail 指针
func (q *LKQueue[T]) Dequeue() T {
	var t T
	for {
		head := load[T](&q.head)
		tail := load[T](&q.tail)
		next := load[T](&head.next)
		if head == load[T](&q.head) { // 检查 head、tail 和 next 是否一致
			if head == tail { // 队列为空，或者 tail 还未到队尾
				if next == nil { // 为空
					return t
				}
				// 将 tail 往队尾移动
				cas(&q.tail, tail, next)
			} else {
				v := next.value
				if cas(&q.head, head, next) {
					return v // 出队完成
				}
			}
		}
	}
}

// 读取节点的值
func load[T any](p *unsafe.Pointer) (n *node[T]) {
	return (*node[T])(atomic.LoadPointer(p))
}

// 原子地修改节点的值
func cas[T any](p *unsafe.Pointer, old, new *node[T]) (ok bool) {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
