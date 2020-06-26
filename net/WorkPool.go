package net

import (
	"github.com/back0893/goTcp/utils"
	"log"
	"sync"
)

var Pool *WorkPool

func init() {
	if utils.GlobalConfig.GetBool("work.Start") {
		utils.GlobalConfig.SetDefault("workBuffer", 5)
		utils.GlobalConfig.SetDefault("workSize", 5)
	}
	Pool = NewWorkPool()
	Pool.Start()
}

type WorkPool struct {
	tasks   chan func()
	start   chan *struct{}
	closed  bool
	once    sync.Once
	workNum int
}

func (w *WorkPool) Close() {
	w.once.Do(func() {
		w.closed = true
		close(w.tasks)
		close(w.start)
	})
}

func NewWorkPool() *WorkPool {
	workBuffer := utils.GlobalConfig.GetInt("workBuffer")
	workSize := utils.GlobalConfig.GetInt("workSize")
	return &WorkPool{
		tasks:   make(chan func(), workBuffer),
		start:   make(chan *struct{}),
		workNum: workSize,
	}
}

func (w *WorkPool) Add(fn func()) {
	if w.closed {
		return
	}
	w.tasks <- fn
}
func (w *WorkPool) Work() {
	defer func() {
		if err := recover(); err != nil {
			//因为当前的携程数量退出了,需要重启
			log.Println("work except exit")
			if !w.closed {
				w.start <- nil
			}
		} else {
			log.Println("work exit")
		}
	}()
	for fn := range w.tasks {
		fn()
	}

}
func (w *WorkPool) listenWorkStart() {
	for _ = range w.start {
		go w.Work()
	}
}
func (w *WorkPool) Start() {
	go w.listenWorkStart()
	for i := 0; i < w.workNum; i++ {
		go w.Work()
	}
}
