package iface

import (
	"sync"
)

type IServer interface {
	AddProtocol(protocol IProtocol)
	AddConEvent(event IConEvent)
	GetConnections() *sync.Map
	StoreCon(connection IConnection)
	DeleteCon(connection IConnection)
	GetCon(id uint32)
	Run()
	Stop()
	Listen()
}
