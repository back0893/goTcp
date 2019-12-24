package net

/**
连接成功的事件
分别是
1.连接成功
2.消息获得
3.连接关闭
*/

type IConEvent interface {
	OnConnect(connection *Connection)
	OnMessage(p IPacket, connection *Connection)
	OnClose(connection *Connection)
}
