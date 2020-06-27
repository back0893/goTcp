package iface

type IProtocol interface {
	Pack(pack IPacket) ([]byte, error)
	UnPack(data []byte, atEOF bool) (advance int, token []byte, err error)
	Decode([]byte) (IPacket, error)
}
