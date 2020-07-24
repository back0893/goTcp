package test

import (
	"context"
	"github.com/back0893/goTcp/net"
	"log"
	"testing"
	"time"
)

func TestTimeWheel(t *testing.T) {
	wheel := net.NewTimingWheel(context.Background())
	wheel.AddTimer(time.Now(), 1*time.Second, func() {
		log.Println("1===1")
	})
	wheel.AddTimer(time.Now(), 2*time.Second, func() {
		log.Println("2===2")
	})
	go wheel.Start()
	time.Sleep(4 * time.Second)
	wheel.CancelAll()
	time.Sleep(4 * time.Second)
}
func TestTimeWheelCancel(t *testing.T) {
	wheel := net.NewTimingWheel(context.Background())
	id := wheel.AddTimer(time.Now(), 1*time.Second, func() {
		log.Println("1===1")
	})
	go wheel.Start()
	time.Sleep(3 * time.Second)
	wheel.Cancel(id)
	time.Sleep(3 * time.Second)
}
