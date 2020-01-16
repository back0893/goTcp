package iface

import (
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
}
