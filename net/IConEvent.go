package net

import "context"

/**
连接成功的事件
分别是
1.连接成功
2.消息获得
3.连接关闭
*/

type IConEvent interface {
	OnConnect(ctx context.Context, connection *Connection)
	OnMessage(ctx context.Context, p IPacket, connection *Connection)
	OnClose(ctx context.Context, connection *Connection)
}
