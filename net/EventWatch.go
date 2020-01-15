package net

import (
	"context"
	"github.com/back0893/goTcp/iface"
)

type EventWatch struct {
	connFunc    []func(ctx context.Context, connection iface.IConnection)
	messageFunc []func(ctx context.Context, packet iface.IPacket, connection iface.IConnection)
	closeFunc   []func(ctx context.Context, connection iface.IConnection)
}

func (event *EventWatch) AddConnect(fn func(ctx context.Context, connection iface.IConnection)) {
	event.connFunc = append(event.connFunc, fn)
}

func (event *EventWatch) Connect(ctx context.Context, connection iface.IConnection) {
	for _, fn := range event.connFunc {
		fn(ctx, connection)
	}
}

func (event *EventWatch) AddMessage(fn func(ctx context.Context, packet iface.IPacket, connection iface.IConnection)) {
	event.messageFunc = append(event.messageFunc, fn)
}

func (event *EventWatch) Message(ctx context.Context, packet iface.IPacket, connection iface.IConnection) {
	for _, fn := range event.messageFunc {
		fn(ctx, packet, connection)
	}
}

func (event *EventWatch) AddClose(fn func(ctx context.Context, connection iface.IConnection)) {
	event.closeFunc = append(event.closeFunc, fn)
}

func (event *EventWatch) Close(ctx context.Context, connection iface.IConnection) {
	for _, fn := range event.closeFunc {
		fn(ctx, connection)
	}
}

func NewEventWatch() iface.IEventWatch {
	return &EventWatch{
		connFunc:    make([]func(ctx context.Context, connection iface.IConnection), 0, 10),
		messageFunc: make([]func(ctx context.Context, packet iface.IPacket, connection iface.IConnection), 0, 10),
		closeFunc:   make([]func(ctx context.Context, connection iface.IConnection), 0, 10),
	}
}
