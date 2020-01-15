package net

import (
	"context"
	"github.com/back0893/goTcp/iface"
)

type Convent struct {
	connFunc    []func(ctx context.Context, connection iface.IConnection)
	messageFunc []func(ctx context.Context, connection iface.IConnection)
	closeFunc   []func(ctx context.Context, connection iface.IConnection)
}

func (event *Convent) AddConnect(fn func(ctx context.Context, connection iface.IConnection)) {
	event.connFunc = append(event.connFunc, fn)
}

func (event *Convent) Connect(ctx context.Context, connection iface.IConnection) {
	for _, fn := range event.connFunc {
		fn(ctx, connection)
	}
}

func (event *Convent) AddMessage(fn func(ctx context.Context, connection iface.IConnection)) {
	event.messageFunc = append(event.messageFunc, fn)
}

func (event *Convent) Message(ctx context.Context, connection iface.IConnection) {
	for _, fn := range event.messageFunc {
		fn(ctx, connection)
	}
}

func (event *Convent) AddClose(fn func(ctx context.Context, connection iface.IConnection)) {
	event.closeFunc = append(event.closeFunc, fn)
}

func (event *Convent) Close(ctx context.Context, connection iface.IConnection) {
	for _, fn := range event.closeFunc {
		fn(ctx, connection)
	}
}
