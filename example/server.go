package main

import (
	"bufio"
	"context"
	"github.com/back0893/goTcp/iface"
	"github.com/back0893/goTcp/net"
	"github.com/back0893/goTcp/utils"
	"log"
)

type event struct{}

func (e event) OnConnect(ctx context.Context, connection iface.IConnection) {
	log.Printf("connect id=>%d", connection.GetId())
}

func (e event) OnMessage(ctx context.Context, p iface.IPacket, connection iface.IConnection) {
	pkt := p.(*packet)
	log.Println(string(pkt.data))
	if err := connection.Write(pkt); err != nil {
		log.Println(err)
	}
}

func (e event) OnClose(ctx context.Context, connection iface.IConnection) {
	log.Printf("close")
}

type protocol struct{}

func (p protocol) UnPack(data []byte, atEOF bool) (advance int, token []byte, err error) {
	return bufio.ScanLines(data, atEOF)
}

func (p protocol) Decode(data []byte) iface.IPacket {
	return &packet{data}
}

func (p protocol) Pack(pack iface.IPacket) ([]byte, error) {
	return pack.Serialize()
}

type packet struct {
	data []byte
}

func (p *packet) Serialize() ([]byte, error) {
	return p.data, nil
}

func main() {
	utils.GlobalConfig.Load("json", "./example/app.json")
	s := net.NewServer()
	s.AddEvent(&event{})
	s.AddProtocol(&protocol{})
	ip := utils.GlobalConfig.GetString("Ip")
	port := utils.GlobalConfig.GetInt("Port")
	net.StartWorkPool()

	s.Listen(ip, port)
}
