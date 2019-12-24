package net

import (
	"bufio"
	"errors"
	"github.com/back0893/goTcp/utils"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Connection struct {
	server            *Server
	conn              *net.TCPConn
	closeOnce         sync.Once //关闭的唯一操作
	closeFlag         int32
	closeChan         chan bool
	packetSendChan    chan IPacket
	packetReceiveChan chan IPacket
	conId             uint32
	buffer            *bufio.Reader //包装tcpConn,方便读取
}

func newConn(conn *net.TCPConn, server *Server, conId uint32) *Connection {
	c := &Connection{
		server:            server,
		conn:              conn,
		conId:             conId,
		closeChan:         make(chan bool),
		packetSendChan:    make(chan IPacket, utils.GlobalConfig.PacketSendChanLimit),
		packetReceiveChan: make(chan IPacket, utils.GlobalConfig.PacketReceiveChanLimit),
		buffer:            bufio.NewReader(conn),
	}
	c.server.ConEvent.OnConnect(c)
	return c
}
func (c *Connection) GetBuffer() *bufio.Reader {
	return c.buffer
}
func (c *Connection) GetId() uint32 {
	return c.conId
}
func (c *Connection) GetServer() *Server {
	return c.server
}
func (c *Connection) Close() {
	//使用sync.Once确保关闭只能被执行一次,
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closeFlag, 1)
		close(c.closeChan)
		close(c.packetSendChan)
		close(c.packetReceiveChan)
		c.server.ConEvent.OnClose(c)
		//连接关闭时,删除.
		c.server.connections.Delete(c.conId)
		c.conn.Close()
	})
}

func (c *Connection) IsClosed() bool {
	return atomic.LoadInt32(&c.closeFlag) == 1
}
func (c *Connection) run() {
	utils.AsyncDo(c.ReadLoop, c.server.waitGroup)
	utils.AsyncDo(c.WriteLoop, c.server.waitGroup)
	utils.AsyncDo(c.HandLoop, c.server.waitGroup)
}
func (c *Connection) ReadLoop() {
	defer func() {
		//如果有错误产生,这里捕获
		//防止整个服务退出
		recover()
		c.Close()
	}()
	for {
		select {
		case <-c.server.exitChan:
			return
		case <-c.closeChan:
			return
		default:

		}
		p, err := c.server.protocol.UnPack(c)
		if err != nil {
			return
		}
		c.packetReceiveChan <- p
	}
}
func (c *Connection) WriteLoop() {
	defer func() {
		recover()
		c.Close()
	}()
	for {
		select {
		case <-c.server.exitChan:
			return
		case <-c.closeChan:
			return
		case p := <-c.packetSendChan:
			if c.IsClosed() {
				return
			}
			raw, err := c.server.protocol.Pack(p)
			if err != nil {
				return
			}
			if _, err := c.conn.Write(raw); err != nil {
				return
			}
		}

	}
}
func (c *Connection) HandLoop() {
	defer func() {
		recover()
		c.Close()
	}()
	for {
		select {
		case <-c.server.exitChan:
			return
		case <-c.closeChan:
			return
		case p := <-c.packetReceiveChan:
			if c.IsClosed() {
				return
			}
			c.server.ConEvent.OnMessage(p, c)
		}
	}
}

func (c *Connection) Write(p IPacket) error {
	return c.AsyncWrite(p, 0)
}
func (c *Connection) AsyncWrite(p IPacket, timeout time.Duration) error {
	if c.IsClosed() {
		return errors.New("关闭")
	}
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if timeout == 0 {
		select {
		case c.packetSendChan <- p:
			return nil
		default:
			return errors.New("发送超时")
		}
	} else {
		select {
		case c.packetSendChan <- p:
			return nil
		case <-c.closeChan:
			return nil
		case <-time.After(timeout):
			return errors.New("发送超时")
		}
	}
}
