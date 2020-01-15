package utils

import "sync"

func AsyncDo(fn func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
}
