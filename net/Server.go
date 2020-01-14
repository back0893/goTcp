package net

import (
	"context"
	"fmt"
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
	protocol    IProtocol
	ConEvent    IConEvent
	connections *sync.Map
	cancel      context.CancelFunc
}

func NewServer(event IConEvent) *Server {
	return &Server{
		waitGroup:   &sync.WaitGroup{},
		acceptChan:  make(chan *net.TCPConn),
		ConEvent:    event,
		connections: &sync.Map{},
	}
}
func (server *Server) AddProtocol(protocol IProtocol) {
	server.protocol = protocol
}
func (server *Server) GetConnections() *sync.Map {
	return server.connections
}
func (server *Server) Run() {
	s := fmt.Sprintf("%s:%d", utils.GlobalConfig.GetString("Ip"), utils.GlobalConfig.GetInt("Port"))
	addr, err := net.ResolveTCPAddr("tcp", s)
	if err != nil {
		log.Print(err)
		return
	}
	listner, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Print(err)
		return
	}
	go server.accept(listner)

	/**
	2020年1月14日 使用context改造
	*/
	ctx, cancel := context.WithCancel(context.Background())
	server.cancel = cancel

	go func() {
		var conId uint32 = 0
		for {
			select {
			case err := <-ctx.Done():
				log.Println(err)
				return
			case conn := <-server.acceptChan:
				conId++
				go func() {
					con := newConn(ctx, conn, server.ConEvent, server.protocol, server.waitGroup, conId)
					server.waitGroup.Add(1)
					server.connections.Store(conId, con)
					con.run()
				}()
			default:
			}
		}
	}()
	log.Printf("tcp监听:%s", s)
}
func (server *Server) accept(listener *net.TCPListener) {
	defer listener.Close()
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}
		server.acceptChan <- conn
	}
}
func (server *Server) Stop() {
	server.cancel()
}

func (server *Server) Listen() {
	server.Run()
	log.Println("接受停止或者ctrl-c停止")
	chSign := make(chan os.Signal)
	signal.Notify(chSign, syscall.SIGINT, syscall.SIGTERM)
	log.Println("接受到信号:", <-chSign)
	server.Stop()
}
