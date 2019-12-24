package net

type IPacket interface {
	Serialize() ([]byte, error)
}
