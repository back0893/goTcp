package iface

type IWorkPool interface {
	Add(func())
	Start()
	Close()
}
