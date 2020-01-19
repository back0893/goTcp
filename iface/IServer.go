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
	WithContextValue(func(ctx context.Context) context.Context) //对于上下文在服务的操作
}
