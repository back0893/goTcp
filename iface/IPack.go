package iface

type IPacket interface {
	Serialize() ([]byte, error)
}
