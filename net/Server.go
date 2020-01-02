package net

import (
	"fmt"
	"github.com/back0893/goTcp/utils"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server struct {
	waitGroup   *sync.WaitGroup //退出时平滑退出
	exitChan    chan bool       //退出时直接关闭,所有监听都会收到
	protocol    IProtocol
	ConEvent    IConEvent
	connections *sync.Map
}

func NewServer(event IConEvent) *Server {
	return &Server{
		waitGroup:   &sync.WaitGroup{},
		exitChan:    make(chan bool),
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

	server.waitGroup.Add(1)
	var conId uint32 = 0
	go func() {
		defer func() {
			listner.Close()
			server.waitGroup.Done()
		}()
		for {
			select {
			case <-server.exitChan:
				return
			default:
			}
			//10秒过期,方便循环
			listner.SetDeadline(time.Now().Add(10 * time.Second))
			conn, err := listner.AcceptTCP()
			if err != nil {
				continue
			}
			conId++
			server.waitGroup.Add(1)
			go func() {
				con := newConn(conn, server, conId)
				server.connections.Store(conId, con)
				con.run()
				server.waitGroup.Done()
			}()
		}
	}()
	log.Printf("tcp监听:%s", s)
}

func (server *Server) Stop() {
	close(server.exitChan)
	server.waitGroup.Wait()
}

func (server *Server) Listen() {
	server.Run()
	log.Println("接受停止或者ctrl-c停止")
	chSign := make(chan os.Signal)
	signal.Notify(chSign, syscall.SIGINT, syscall.SIGTERM)
	log.Println("接受到信号:", <-chSign)
	server.Stop()
}
