package utils

import "sync/atomic"

/**
原子的唯一值
*/

type AtomicInt64 struct {
	num int64
}

func NewAtomicInt64(initialValue int64) *AtomicInt64 {
	return &AtomicInt64{
		num: initialValue,
	}
}

func (a *AtomicInt64) Get() int64 {
	return atomic.LoadInt64(&a.num)
}
func (a *AtomicInt64) Store(value int64) {
	atomic.StoreInt64(&a.num, value)
}
func (a *AtomicInt64) Add(increment int64) (int64, bool) {
	current := a.Get()
	return current + increment, a.CompareAndSet(current, current+increment)
}
func (a *AtomicInt64) Sub(sub int64) (int64, bool) {
	current := a.Get()
	return current - sub, a.CompareAndSet(current, current-sub)
}
func (a *AtomicInt64) CompareAndSet(except, update int64) bool {
	return atomic.CompareAndSwapInt64(&a.num, except, update)
}
