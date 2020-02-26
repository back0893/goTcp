package net

import (
	"context"
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
	acceptChan   chan *net.TCPConn //接受socket使用协成
	waitGroup    *sync.WaitGroup
	protocol     iface.IProtocol
	ConEvent     iface.IEventWatch
	connections  *sync.Map
	ctxCancel    context.CancelFunc
	ctx          context.Context
	listener     *net.TCPListener
	contextValue func(ctx context.Context) context.Context
}

func (s *Server) WithContextValue(fn func(ctx context.Context) context.Context) {
	s.contextValue = fn
}
func (s *Server) context() {
	/**
	2020年1月14日 使用context改造
	*/
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	if s.contextValue != nil {
		s.ctx = s.contextValue(s.ctx)
	}

}
func NewServer() *Server {
	s := &Server{
		waitGroup:   &sync.WaitGroup{},
		acceptChan:  make(chan *net.TCPConn),
		connections: &sync.Map{},
		ConEvent:    NewEventWatch(),
	}
	/**
	在事件上新增开始和结束的事件
	为了给server的conns新增和删除链接
	*/
	s.ConEvent.AddConnect(func(ctx context.Context, connection iface.IConnection) {
		s.connections.Store(connection.GetId(), connection)
	})
	s.ConEvent.AddClose(func(ctx context.Context, connection iface.IConnection) {
		s.connections.Delete(connection.GetId())
	})
	return s
}
func (s *Server) AddProtocol(protocol iface.IProtocol) {
	s.protocol = protocol
}
func (s *Server) GetConnections() *sync.Map {
	return s.connections
}
func (s *Server) Run() {
	str := net.JoinHostPort(utils.GlobalConfig.GetString("Ip"), utils.GlobalConfig.GetString("Port"))
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

	s.context()

	go func() {
		var conId uint32 = 0
		for {
			select {
			case <-s.ctx.Done():
				return
			case conn := <-s.acceptChan:
				conId++
				con := NewConn(s.ctx, conn, s.waitGroup, s.ConEvent, s.protocol, conId)
				go con.Run()
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

func (s *Server) AddEvent(event iface.IEvent) {
	s.ConEvent.AddConnect(event.OnConnect)
	s.ConEvent.AddMessage(event.OnMessage)
	s.ConEvent.AddClose(event.OnClose)
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
