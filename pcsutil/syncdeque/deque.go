// Package syncdeque 提供一个线程安全的双端队列, 用于替代之前依赖的
// github.com/oleiade/lane.Deque. 仓库内只用到 Append / Shift / Size
// 这三个方法, 这里就只暴露这些, 没必要再现一个全套 deque API.
package syncdeque

import (
	"container/list"
	"sync"
)

// Deque 是一个并发安全、容量不限的 FIFO/双端队列.
// 与 oleiade/lane 不同, 这里用泛型, 调用方不需要 type assertion.
type Deque[T any] struct {
	mu sync.Mutex
	l  list.List
}

// New 创建一个空 Deque.
func New[T any]() *Deque[T] {
	return &Deque[T]{}
}

// Append 将元素放到队尾, O(1).
func (d *Deque[T]) Append(v T) {
	d.mu.Lock()
	d.l.PushBack(v)
	d.mu.Unlock()
}

// Shift 从队首弹出一个元素. 若队列为空则返回 (零值, false).
func (d *Deque[T]) Shift() (T, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	front := d.l.Front()
	var zero T
	if front == nil {
		return zero, false
	}
	d.l.Remove(front)
	return front.Value.(T), true
}

// Size 返回当前元素个数.
func (d *Deque[T]) Size() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.l.Len()
}
