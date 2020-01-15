package iface

import (
	"context"
)

/**
连接成功的事件
分别是
1.连接成功
2.消息获得
3.连接关闭
*/

type IEventWatch interface {
	AddConnect(func(ctx context.Context, connection IConnection))
	Connect(ctx context.Context, connection IConnection)
	AddMessage(func(ctx context.Context, packet IPacket, connection IConnection))
	Message(ctx context.Context, packet IPacket, connection IConnection)
	AddClose(func(ctx context.Context, connection IConnection))
	Close(ctx context.Context, connection IConnection)
}
