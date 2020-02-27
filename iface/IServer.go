package iface

import (
	"context"
	"sync"
)

type IServer interface {
	AddProtocol(protocol IProtocol)
	GetConnections() *sync.Map
	StoreCon(connection IConnection)
	DeleteCon(connection IConnection)
	GetCon(id uint32) (IConnection, bool)
	Run()
	Stop()
	Listen()
	AddEvent(event IEvent)
	GetContext() context.Context // 返回上下文.
}
