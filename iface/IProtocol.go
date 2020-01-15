package iface

type IProtocol interface {
	Pack(pack IPacket) ([]byte, error)
	UnPack(conn IConnection) (IPacket, error)
}
