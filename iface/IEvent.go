package iface

import "context"

type IEvent interface {
	OnConnect(ctx context.Context, connection IConnection)
	OnMessage(ctx context.Context, packet IPacket, connection IConnection)
	OnClose(ctx context.Context, connection IConnection)
}
