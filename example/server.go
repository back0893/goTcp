package main

import (
	"context"
	"github.com/back0893/goTcp/iface"
	"github.com/back0893/goTcp/net"
	"github.com/back0893/goTcp/utils"
	"log"
)

type event struct{}

func (e event) OnConnect(ctx context.Context, connection iface.IConnection) {
	log.Printf("connect")
}

func (e event) OnMessage(ctx context.Context, p iface.IPacket, connection iface.IConnection) {
	pkt := p.(*packet)
	log.Println(string(pkt.data))
	connection.Write(pkt)
}

func (e event) OnClose(ctx context.Context, connection iface.IConnection) {
	log.Printf("close")
}

type protocol struct{}

func (p protocol) Pack(pack iface.IPacket) ([]byte, error) {
	return pack.Serialize()
}

func (p protocol) UnPack(conn iface.IConnection) (iface.IPacket, error) {
	l, err := conn.GetBuffer().ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	return &packet{data: l}, nil
}

type packet struct {
	data []byte
}

func (p *packet) Serialize() ([]byte, error) {
	return p.data, nil
}

func main() {
	utils.GlobalConfig.Load("json", "./app.json")
	s := net.NewServer()
	s.AddEvent(&event{})
	s.AddProtocol(&protocol{})
	s.Listen()
}
