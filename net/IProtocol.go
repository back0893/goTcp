package net

type IProtocol interface {
	Pack(pack IPacket) ([]byte, error)
	UnPack(conn *Connection) (IPacket, error)
}
