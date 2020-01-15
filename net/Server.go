package net

import (
	"context"
	"fmt"
	"github.com/back0893/goTcp/iface"
	"github.com/back0893/goTcp/utils"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Server struct {
	acceptChan  chan *net.TCPConn //接受socket使用协成
	waitGroup   *sync.WaitGroup
	protocol    iface.IProtocol
	ConEvent    iface.IConEvent
	connections *sync.Map
	ctxCancel   context.CancelFunc
	ctx         context.Context
	listener    *net.TCPListener
}

func NewServer() *Server {
	return &Server{
		waitGroup:   &sync.WaitGroup{},
		acceptChan:  make(chan *net.TCPConn),
		connections: &sync.Map{},
	}
}
func (s *Server) AddProtocol(protocol iface.IProtocol) {
	s.protocol = protocol
}
func (s *Server) GetConnections() *sync.Map {
	return s.connections
}
func (s *Server) Run() {
	str := fmt.Sprintf("%s:%d", utils.GlobalConfig.GetString("Ip"), utils.GlobalConfig.GetInt("Port"))
	addr, err := net.ResolveTCPAddr("tcp", str)
	if err != nil {
		log.Print(err)
		return
	}
	s.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		log.Print(err)
		return
	}
	go s.accept()

	/**
	2020年1月14日 使用context改造
	*/
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	go func() {
		var conId uint32 = 0
		for {
			select {
			case <-s.ctx.Done():
				return
			case conn := <-s.acceptChan:
				conId++
				con := newConn(s.ctx, conn, s.waitGroup, s.ConEvent, s.protocol, conId)
				s.connections.Store(conId, con)
				go con.run()
			}
		}
	}()
	log.Printf("tcp监听:%s", str)
}
func (s *Server) accept() {
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()
	for {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			//这里如果服务器停止监听了会返回err
			if err := s.ctx.Err(); err != nil {
				log.Println("accept return ")
				return
			}
			continue
		}
		s.acceptChan <- conn
	}
}
func (s *Server) Stop() {
	s.listener.Close()
	s.ctxCancel()
	s.waitGroup.Wait()
}

func (s *Server) Listen() {
	s.Run()
	log.Println("接受停止或者ctrl-c停止")
	chSign := make(chan os.Signal)
	signal.Notify(chSign, syscall.SIGINT, syscall.SIGTERM)
	log.Println("接受到信号:", <-chSign)
	s.Stop()
}

func (s *Server) AddConEvent(event iface.IConEvent) {
	/**
	这里需要为实践新增连接成功和连接关闭的事件
	*/
	s.ConEvent = event
}

func (s *Server) StoreCon(connection iface.IConnection) {
	s.connections.Store(connection.GetId(), connection)
}

func (s *Server) DeleteCon(connection iface.IConnection) {
	s.connections.Delete(connection.GetId())
}

func (s *Server) GetCon(id uint32) (con iface.IConnection, ok bool) {
	val, ok := s.connections.Load(id)
	if ok {
		return val.(iface.IConnection), ok
	}
	return nil, ok
}
