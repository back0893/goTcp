package iface

import (
	"bufio"
	"net"
	"sync"
	"time"
)

type IConnection interface {
	GetId() uint32
	GetRawCon() net.Conn
	GetBuffer() *bufio.Reader
	Write(p IPacket) error
	AsyncWrite(p IPacket, timeout time.Duration) error
	GetExtraData(key interface{}) (interface{}, bool)
	SetExtraData(key interface{}, value interface{})
	GetExtraMap() *sync.Map
	Close()
	IsClosed() bool
	Run()
}
