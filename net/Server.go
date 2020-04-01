package net

import (
	"context"
	"fmt"
	"github.com/back0893/goTcp/iface"
	errors2 "github.com/back0893/goTcp/vendor/github.com/pkg/errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

type Server struct {
	acceptChan  chan *net.TCPConn //接受socket使用协成
	waitGroup   *sync.WaitGroup
	protocol    iface.IProtocol
	ConEvent    iface.IEventWatch
	connections *sync.Map
	ctxCancel   context.CancelFunc
	ctx         context.Context
	listener    *net.TCPListener
}

func (s *Server) GetContext() context.Context {
	return s.ctx
}
func (s *Server) SetContext(ctx context.Context) {
	s.ctx, s.ctxCancel = context.WithCancel(ctx)
}
func NewServer() *Server {
	s := &Server{
		waitGroup:   &sync.WaitGroup{},
		acceptChan:  make(chan *net.TCPConn),
		connections: &sync.Map{},
		ConEvent:    NewEventWatch(),
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
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
func (s *Server) Run(ip string, port int) {
	str := net.JoinHostPort(ip, strconv.Itoa(port))
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

func (s *Server) Listen(ip string, port int) {
	//记录异常退出
	defer func() {
		if err := recover(); err != nil {
			stackerr := errors2.WithStack(errors2.New(fmt.Sprintln(err)))
			log.Println(fmt.Printf("%+v", stackerr))
		}
	}()
	s.Run(ip, port)
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
