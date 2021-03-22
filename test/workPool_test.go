package test

import (
	"errors"
	"fmt"
	"github.com/back0893/goTcp/utils"
	"log"
	"sync"
	"testing"
)

type test struct {
	Num int
}

func init() {
	utils.GlobalConfig.Load("json", "./app.json")
	utils.GlobalConfig.SetDefault("work.buffer", 5)
	utils.GlobalConfig.SetDefault("work.size", 5)
}
func TestWorkPool(t *testing.T) {
	pool := utils.NewWorkPool()
	pool.Start()
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		pool.Add(func(id int) func() {
			return func() {
				log.Printf("asdadsa====>%d", id)
				wg.Done()
			}
		}(i))
	}
	wg.Wait()
	pool.Close()
}
func TestWorkPoolExit(t *testing.T) {
	pool := utils.NewWorkPool()
	pool.Start()
	for i := 0; i < 5; i++ {
		pool.Add(func() {
			panic(errors.New("模拟退出"))
		})
	}
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		pool.Add(func(id int) func() {
			return func() {
				fmt.Printf("asdadsa====>%d\n", id)
				wg.Done()
			}
		}(i))
	}
	wg.Done()
	pool.Close()
}

func TestWorkPoolClose(t *testing.T) {
	pool := utils.NewWorkPool()
	pool.Start()
	pool.Close()
	pool.Add(func() {
		fmt.Println(123)
	})
}
func TestWorkPoolWithStruct(t *testing.T) {
	pool := utils.NewWorkPool()
	pool.Start()
	wg := sync.WaitGroup{}
	wg.Add(10)
	tmp := make(chan *test)
	go func() {
		for i := 0; i < 10; i++ {
			tmp <- &test{Num: i}
		}
	}()
	go func() {
		for t := range tmp {
			pool.Add(func(t *test) func() {
				return func() {
					log.Printf("asdadsa====>%d", t.Num)
					wg.Done()
				}
			}(t))
		}

	}()
	wg.Wait()
	pool.Close()
}
