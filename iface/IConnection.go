package iface

type IConnection interface {
	ReadLoop()
	WriteLoop()
	HandLoop()
	GetId() uint32
}
