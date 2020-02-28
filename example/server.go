package main

import (
	"bytes"
	"context"
	"encoding/binary"
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
	ip := utils.GlobalConfig.GetString("Ip")
	port := utils.GlobalConfig.GetInt("Port")

	innerServer(s)

	s.Listen(ip, port)
}

type inpacket struct {
	data []byte
	id   uint32
}

func (p *inpacket) Serialize() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, p.id)
	binary.Write(buffer, binary.BigEndian, p.data)
	return buffer.Bytes(), nil
}

type inevent struct{}

func (e inevent) OnConnect(ctx context.Context, connection iface.IConnection) {
	log.Printf("connect")
}

func (e inevent) OnMessage(ctx context.Context, p iface.IPacket, connection iface.IConnection) {
	pkt := p.(*inpacket)
	s := ctx.Value("pserver").(*net.Server)
	cc, ok := s.GetCon(pkt.id)
	if ok == false {
		log.Printf("id=>%d对应的连接不存在", pkt.id)
		return
	}
	cc.Write(&packet{data: pkt.data})

}

func (e inevent) OnClose(ctx context.Context, connection iface.IConnection) {
	log.Printf("close")
}

type inprotocol struct{}

func (p inprotocol) Pack(pack iface.IPacket) ([]byte, error) {
	return pack.Serialize()
}

func (p inprotocol) UnPack(conn iface.IConnection) (iface.IPacket, error) {
	l, err := conn.GetBuffer().ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	return &inpacket{
		id:   1,
		data: l,
	}, nil
}
func innerServer(s iface.IServer) {
	s2 := net.NewServer()
	s2.AddEvent(&inevent{})
	s2.AddProtocol(&inprotocol{})
	ctx := context.WithValue(s.GetContext(), "pserver", s)
	s2.SetContext(ctx)
	go s2.Listen("0.0.0.0", 10087)
}
