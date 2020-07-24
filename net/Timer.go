package net

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

/**
定时器,观察tao的实现
*/

var (
	timersId *AtomicInt64
)

func init() {
	timersId = NewAtomicInt64(0)
}

const (
	tickPeriod time.Duration = 500 * time.Millisecond
)

type TimerType struct {
	id         int64
	expiration time.Time     //下次执行时间
	interval   time.Duration //执行间隔
	timeout    func()        //执行函数
	index      int           //heap中的索引
}

//IsRepeat 判断是否需要重复执行
func (t *TimerType) IsRepeat() bool {
	return t.interval > 0
}

//Update 更新下次的执行时间
func (t *TimerType) Update() {
	t.expiration = t.expiration.Add(t.interval)
}
func NewTimerType(when time.Time, interval time.Duration, timeout func()) *TimerType {
	for {
		id, ok := timersId.Add(1)
		if ok {
			return &TimerType{
				id:         id,
				expiration: when,
				interval:   interval,
				timeout:    timeout,
				index:      0,
			}
		}
	}
}

type TimerHeapType []*TimerType

func (t TimerHeapType) Len() int {
	return len(t)
}

func (t TimerHeapType) Less(i, j int) bool {
	return t[i].expiration.Before(t[j].expiration)
}

func (t TimerHeapType) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
	t[i].index = i
	t[j].index = j
}

func (t *TimerHeapType) Push(x interface{}) {
	timer, ok := x.(*TimerType)
	if !ok {
		return
	}
	timer.index = len(*t)
	*t = append(*t, timer)
}
func (t *TimerHeapType) getIndexById(id int64) int {
	for index, timerType := range *t {
		if timerType.id == id {
			return index
		}
	}
	return -1
}
func (t *TimerHeapType) Pop() interface{} {
	n := len(*t)
	timer := (*t)[n-1]
	timer.index = -1
	*t = (*t)[:n-1]
	return timer
}

type TimingWheel struct {
	timers     *TimerHeapType
	ticker     *time.Ticker
	ctx        context.Context
	cancel     context.CancelFunc
	once       sync.Once
	cancelChan chan int64
}

func NewTimingWheel(ctx context.Context) *TimingWheel {
	h := make(TimerHeapType, 0)
	wheel := &TimingWheel{
		timers:     &h,
		ticker:     time.NewTicker(tickPeriod),
		once:       sync.Once{},
		cancelChan: make(chan int64),
	}
	wheel.ctx, wheel.cancel = context.WithCancel(ctx)
	return wheel
}

func (wheel *TimingWheel) AddTimer(when time.Time, interval time.Duration, timeout func()) int64 {
	timer := NewTimerType(when, interval, timeout)
	heap.Push(wheel.timers, timer)
	return timer.id
}
func (wheel *TimingWheel) Stop() {
	wheel.once.Do(func() {
		wheel.timers = nil
		wheel.ticker.Stop()
	})
}
func (wheel *TimingWheel) GetExpired() []*TimerType {
	expired := make([]*TimerType, 0)
	now := time.Now()
	for wheel.timers.Len() > 0 {
		timerType := heap.Pop(wheel.timers).(*TimerType)
		elapese := now.Sub(timerType.expiration).Seconds()
		if elapese > 0.0 {
			expired = append(expired, timerType)
		} else {
			heap.Push(wheel.timers, timerType)
			break
		}
	}
	return expired
}
func (wheel *TimingWheel) Update(times []*TimerType) {
	for _, t := range times {
		if t.IsRepeat() {
			t.Update()
			heap.Push(wheel.timers, t)
		}
	}
}

func (wheel *TimingWheel) Cancel(id int64) {
	wheel.cancelChan <- id
}
func (wheel *TimingWheel) Start() {
	for {
		select {
		case <-wheel.ctx.Done():
			wheel.Stop()
			return
		case id := <-wheel.cancelChan:
			//-1 未特殊的停止所有的定时器
			if id == -1 {
				*wheel.timers = (*wheel.timers)[:0]
			} else {
				index := wheel.timers.getIndexById(id)
				if index != -1 {
					heap.Remove(wheel.timers, index)
				}
			}
		case <-wheel.ticker.C:
			timers := wheel.GetExpired()
			for _, timeType := range timers {
				if Pool != nil {
					Pool.Add(timeType.timeout)
				} else {
					go timeType.timeout()
				}
			}
			wheel.Update(timers)
		}
	}
}

func (wheel *TimingWheel) CancelAll() {
	wheel.cancelChan <- -1
}
