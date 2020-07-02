package net

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/back0893/goTcp/iface"
	"github.com/back0893/goTcp/utils"
	errors2 "github.com/pkg/errors"
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
	packetSendChan    chan iface.IPacket
	packetReceiveChan chan iface.IPacket
	conId             uint32
	event             iface.IEventWatch
	protocol          iface.IProtocol
	cancelFunc        context.CancelFunc
}

func NewConn(ctx context.Context, conn *net.TCPConn, wg *sync.WaitGroup, event iface.IEventWatch, protocol iface.IProtocol, conId uint32) *Connection {
	c := &Connection{
		conn:              conn,
		conId:             conId,
		packetSendChan:    make(chan iface.IPacket, utils.GlobalConfig.GetInt("PacketSendChanLimit")),
		packetReceiveChan: make(chan iface.IPacket, utils.GlobalConfig.GetInt("packetReceiveChan")),
		extraData:         &sync.Map{},
		wg:                wg,
		event:             event,
		protocol:          protocol,
	}
	c.wg.Add(1)
	c.ctx, c.cancelFunc = context.WithCancel(ctx)
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
func (c *Connection) GetId() uint32 {
	return c.conId
}
func (c *Connection) Close() {
	//使用sync.Once确保关闭只能被执行一次,
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closeFlag, 1)
		//最先执行执行的关闭
		c.event.Close(c.ctx, c)

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
func (c *Connection) Run() {
	c.event.Connect(c.ctx, c)
	utils.AsyncDo(c.readLoop, c.wg)
	utils.AsyncDo(c.writeLoop, c.wg)
	utils.AsyncDo(c.handLoop, c.wg)
}
func (c *Connection) readLoop() {
	defer func() {
		//如果有错误产生,这里捕获
		//防止整个服务退出
		if err := recover(); err != nil {
			stackerr := errors2.WithStack(errors2.New(fmt.Sprintln(err)))
			log.Printf("%+v", stackerr)
		}
		c.Close()
	}()

	bc := make(chan []byte, 10)
	go func() {
		defer log.Println("close bc.....")
		reader := bufio.NewScanner(c.conn)
		reader.Split(c.protocol.UnPack)
		for reader.Scan() {
			bc <- reader.Bytes()
		}
		close(bc)
	}()
	for {
		select {
		case <-c.ctx.Done():
		case raw, ok := <-bc:
			if !ok {
				return
			}
			if pkt, err := c.protocol.Decode(raw); err != nil {
				//解析失败
				log.Printf("解析失败,%s", err.Error())
			} else {
				c.packetReceiveChan <- pkt
			}
		}
	}
}
func (c *Connection) writeLoop() {
	defer func() {
		if err := recover(); err != nil {
			stackerr := errors2.WithStack(errors2.New(fmt.Sprintln(err)))
			log.Printf("%+v", stackerr)
		}
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
func (c *Connection) handLoop() {
	defer func() {
		if err := recover(); err != nil {
			stackerr := errors2.WithStack(errors2.New(fmt.Sprintln(err)))
			log.Printf("%+v", stackerr)
		}
		c.Close()
	}()
	for {
		work := utils.GlobalConfig.GetBool("work.start")
		select {
		case <-c.ctx.Done():
			return
		case p := <-c.packetReceiveChan:
			if c.IsClosed() {
				return
			}
			if work {
				Pool.Add(func(ctx context.Context, packet iface.IPacket, connection iface.IConnection) func() {
					return func() {
						c.event.Message(ctx, packet, connection)
					}
				}(c.ctx, p, c))
			} else {
				c.event.Message(c.ctx, p, c)
			}
		}
	}
}

func (c *Connection) Write(p iface.IPacket) error {
	return c.AsyncWrite(p, 0)
}
func (c *Connection) AsyncWrite(p iface.IPacket, timeout time.Duration) error {
	if c.IsClosed() {
		return errors.New("关闭")
	}
	defer func() {
		if err := recover(); err != nil {
			stackerr := errors2.WithStack(errors2.New(fmt.Sprintln(err)))
			log.Printf("%+v", stackerr)
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

func (c *Connection) GetRawCon() net.Conn {
	return c.conn
}
