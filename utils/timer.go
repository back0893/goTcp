package utils

import (
	"container/heap"
	"context"
	"log"
	"sync"
	"time"

	"github.com/back0893/goTcp/net"
)

/**
定时器,定时轮训
*/
const (
	tickPeriod time.Duration = 500 * time.Millisecond
)

var (
	timersId    *AtomicInt64
	timingWheel *TimingWheel
)

func GetTimingWheel() *TimingWheel {
	return timingWheel
}
func InitTimingWheel(ctx context.Context) {
	if timingWheel != nil {
		log.Println("全局定时器已经启动,请不要重复启动")
		return
	}
	timingWheel = NewTimingWheel(ctx)
	go timingWheel.Start()
}
func AddTimer(interval time.Duration, fn func()) int64 {
	return timingWheel.AddTimer(time.Now(), interval, fn)
}
func TimerAt(when time.Time, fn func()) int64 {
	return timingWheel.AddTimer(when, 0, fn)
}
func init() {
	timersId = NewAtomicInt64(0)
}
func CancelTimer(id int64) {
	timingWheel.Cancel(id)
}

/**
定时器需要依据时间从近到远被排列
参考了github.com/leesper/tao的实现
*/
type TimerHeapType []*TimerType

func (t TimerHeapType) Len() int {
	return len(t)
}
func (t TimerHeapType) Less(i, j int) bool {
	return t[j].expiration.Sub(t[i].expiration) > 0
}

func (t TimerHeapType) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
	t[i].index = i
	t[j].index = j
}

func (t *TimerHeapType) Push(x interface{}) {
	//这里使用指针是因为如果长度不足,会重新生成新的数组,这回启用的新的地址.
	//这回导致新的地址无法被使用...
	timer, ok := x.(*TimerType)
	if ok == false {
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
	old := *t //复制一份
	n := len(old)
	timer := old[n-1]
	timer.index = -1
	*t = old[:n-1] //覆盖原本
	return timer
}

type TimerType struct {
	id         int64
	expiration time.Time     //下次执行的时间
	interval   time.Duration //执行间隔
	timeout    func()        //执行的方法
	index      int           //在heap中的索引
}

func (t *TimerType) IsRepeat() bool {
	return t.interval > 0
}
func (t *TimerType) Update() {
	//更新下次执行时间
	t.expiration = t.expiration.Add(t.interval)
}
func NewTimerType(when time.Time, interval time.Duration, timeout func()) *TimerType {
	var id int64
	var ok bool
	for {
		id, ok = timersId.Add(1)
		if ok {
			break
		}
	}

	return &TimerType{
		id:         id,
		expiration: when.Add(interval),
		interval:   interval,
		timeout:    timeout,
		index:      0,
	}
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
		ticker:     time.NewTicker(tickPeriod),
		timers:     &h,
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

//获得到期的任务
func (wheel *TimingWheel) GetExpired() []*TimerType {
	expired := make([]*TimerType, 0)
	now := time.Now()
	for wheel.timers.Len() > 0 {
		timerType := heap.Pop(wheel.timers).(*TimerType)
		elapsed := now.Sub(timerType.expiration).Seconds()
		if elapsed > 0.0 {
			expired = append(expired, timerType)
			continue
		} else {
			heap.Push(wheel.timers, timerType)
			//因为heap是排序后的,所以这后面应该没有过期..
			break
		}
	}
	return expired
}

/**
更新到期的任务.
*/
func (wheel *TimingWheel) Update(timers []*TimerType) {
	for _, t := range timers {
		if t.IsRepeat() {
			t.Update()
			heap.Push(wheel.timers, t)
		}
	}
}
func (wheel *TimingWheel) Cancel(id int64) {
	//这里如果直接使用heap会出现数据正在执行时,没有的
	//使用chan发送发
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
				*wheel.timers = (*wheel.timers)[0:0]
			} else {
				index := wheel.timers.getIndexById(id)
				if index > -1 {
					heap.Remove(wheel.timers, index)
				}
			}
		case <-wheel.ticker.C:
			timers := wheel.GetExpired()
			for _, timerType := range timers {
				//这里如果使用一个工作次,避免堵塞
				//调度,如果使用workPool可能出现延迟执行的情况
				//定时任务放到工作池中,如果工作池么有启动
				//使用携程
				if net.Pool != nil {
					net.Pool.Add(timerType.timeout)
				} else {
					go timerType.timeout()
				}
			}
			wheel.Update(timers)
		}
	}
}

func (wheel *TimingWheel) CancelAll() {
	wheel.cancelChan <- (-1)
}
