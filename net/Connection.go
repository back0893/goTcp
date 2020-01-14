package net

import (
	"bufio"
	"context"
	"errors"
	"github.com/back0893/goTcp/utils"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Connection struct {
	ctx               context.Context
	conn              *net.TCPConn
	wg                *sync.WaitGroup
	extraData         *sync.Map //连接保存额外信息
	closeOnce         sync.Once //关闭的唯一操作
	closeFlag         int32
	packetSendChan    chan IPacket
	packetReceiveChan chan IPacket
	conId             uint32
	buffer            *bufio.Reader //包装tcpConn,方便读取
	event             IConEvent
	protocol          IProtocol
	cancelFunc        context.CancelFunc
}

func newConn(ctx context.Context, conn *net.TCPConn, event IConEvent, protocol IProtocol, wg *sync.WaitGroup, conId uint32) *Connection {
	c := &Connection{
		conn:              conn,
		conId:             conId,
		wg:                wg,
		packetSendChan:    make(chan IPacket, utils.GlobalConfig.GetInt("PacketSendChanLimit")),
		packetReceiveChan: make(chan IPacket, utils.GlobalConfig.GetInt("packetReceiveChan")),
		buffer:            bufio.NewReader(conn),
		extraData:         &sync.Map{},
		protocol:          protocol,
	}
	ctx, cancel := context.WithCancel(ctx)
	c.ctx = ctx
	c.cancelFunc = cancel
	c.event.OnConnect(ctx, c)
	return c
}
func (c *Connection) GetExtraData(key interface{}) (interface{}, bool) {
	return c.extraData.Load(key)
}
func (c *Connection) GetExtraMap() *sync.Map {
	return c.extraData
}
func (c *Connection) SetExtraData(key interface{}, value interface{}) {
	c.extraData.Store(key, value)
}
func (c *Connection) GetBuffer() *bufio.Reader {
	return c.buffer
}
func (c *Connection) GetId() uint32 {
	return c.conId
}
func (c *Connection) Close() {
	//使用sync.Once确保关闭只能被执行一次,
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closeFlag, 1)
		//最先执行执行的关闭
		c.event.OnClose(c.ctx, c)

		c.wg.Done()
		c.cancelFunc()
		close(c.packetSendChan)
		close(c.packetReceiveChan)
		c.conn.Close()
	})
}

func (c *Connection) IsClosed() bool {
	return atomic.LoadInt32(&c.closeFlag) == 1
}
func (c *Connection) run() {
	go c.ReadLoop()
	go c.WriteLoop()
	go c.HandLoop()
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
		case <-c.ctx.Done():
			return
		default:

		}
		p, err := c.protocol.UnPack(c)
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
		case <-c.ctx.Done():
			return
		case p := <-c.packetSendChan:
			if c.IsClosed() {
				return
			}
			raw, err := c.protocol.Pack(p)
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
		case <-c.ctx.Done():
			return
		case p := <-c.packetReceiveChan:
			if c.IsClosed() {
				return
			}
			c.event.OnMessage(c.ctx, p, c)
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
		case <-time.After(timeout):
			return errors.New("发送超时")
		}
	}
}
