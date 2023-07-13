package network

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/ssp/msg"
	"github.com/ssp/util"
)

type MsgCmd uint32

const (
	HeartbeatMsgCmd MsgCmd = 5
	RpcMsgCmd       MsgCmd = 11
	FlowMsgCmd      MsgCmd = 12
)

type connectonFlag uint8

const (
	connectionOpenFlag  connectonFlag = 1
	connectionCloseFlag connectonFlag = 2
)

type Connection struct {
	// 底层网络连接
	conn net.Conn

	// Channel ID 生成器
	channelIdGenerator *util.Id

	// Rpc ID 生成器
	requestIdGenerator *util.Id

	// 写缓存
	writerBuff chan []byte

	// 通道集合
	channels map[uint32]*Channel

	// 响应处理集合
	promises map[uint32]*RpcPromise

	// 连接状态
	flag connectonFlag

	// 读写锁，控制对 channels 字段的并发读写
	chMutex sync.RWMutex

	// 读写锁，控制对 flag 字段的并发读写
	flagMutex sync.RWMutex

	//读写锁，控制对 promises 字段的并发读写
	promiseMutex sync.RWMutex
}

func NewConnection(conn net.Conn) *Connection {

	connection := new(Connection)
	connection.conn = conn
	connection.channelIdGenerator = util.NewId(0)
	connection.requestIdGenerator = util.NewId(0)
	connection.writerBuff = make(chan []byte, 1024)

	connection.channels = map[uint32]*Channel{}
	connection.promises = map[uint32]*RpcPromise{}

	connection.flag = connectionOpenFlag

	return connection
}

func (c *Connection) Read() {
	defer util.Trace("", "Client Read")()

	ctx := context.Background()

	reader := bufio.NewReader(c.conn)
	//读消息
	for {
		data, err := msg.Decode(reader)
		m := &msg.Msg{}

		if err == nil && len(data) > 0 {
			err := proto.Unmarshal(data, m)
			if err != nil {
				log.Printf("Close connection:%s:%s \n", c.conn.RemoteAddr(), err.Error())
				c.Close()

				break
			}
		}
		if err != nil {
			log.Printf("Close connection:%s:%s \n", c.conn.RemoteAddr(), err.Error())
			c.Close()

			break
		}

		cmd := MsgCmd(m.Cmd)
		switch cmd {
		case RpcMsgCmd:
			go c.RpcProcess(ctx, m.Data)
		case FlowMsgCmd:
			// 写入channel
			c.Flow(m)
		case HeartbeatMsgCmd:
		default:

		}

	}

}

func (c *Connection) Close() {
	c.flagMutex.Lock()
	c.flag = connectionCloseFlag
	c.flagMutex.Unlock()

	log.Printf("Start close connection .....")
	// 关闭 写缓存
	close(c.writerBuff)

	// 关闭通道
	for id, ch := range c.channels {

		log.Printf("Close channel: %d \n", id)
		ch.Close()
	}

	c.channels = nil
	c.promises = nil

	log.Printf("End close connection.....")
}

func (c *Connection) Write() {
	defer util.Trace("", "Client Write")()

	for data := range c.writerBuff {
		c.conn.Write(data)
	}

}

func (c *Connection) WriteBytes(data []byte) error {

	c.flagMutex.Lock()

	if c.flag == connectionCloseFlag {
		log.Printf("Cann't write data,because connection was closed!\n")

		return errors.New("Connection was closed!")
	}

	c.flagMutex.Unlock()

	c.writerBuff <- data

	return nil
}

func (c *Connection) ApplyChannel() *Channel {

	// 申请一个唯一的通道 id
	id := c.channelIdGenerator.IncrementAndGet()
	channel := NewChannel(id, c)

	c.chMutex.Lock()

	c.channels[id] = channel

	c.chMutex.Unlock()

	return channel

}

func (c *Connection) RegChannel(channelId uint32, channel *Channel) bool {

	c.chMutex.Lock()

	c.channels[channelId] = channel

	c.chMutex.Unlock()

	return true
}

func (c *Connection) RemoveChannel(channelId uint32) bool {

	c.chMutex.Lock()

	delete(c.channels, channelId)

	c.chMutex.Unlock()

	return true
}

func (c *Connection) Flow(msg *msg.Msg) {

	id := msg.Id

	if channel, ok := c.channels[id]; ok {
		channel.AppendReadBuff(msg.Data)
	}
}

func (c *Connection) RegPromise(requestId uint32, promise *RpcPromise) bool {

	c.promiseMutex.Lock()
	c.promises[requestId] = promise
	c.promiseMutex.Unlock()

	return true
}

func (c *Connection) PromiseProcess(result *msg.RpcMsg) {

	requestId := result.Id
	if promise, ok := c.promises[requestId]; ok {
		promise.Set(result)
	}
}

func (c *Connection) RpcProcess(ctx context.Context, message []byte) error {
	rpcMsg := &msg.RpcMsg{}

	if len(message) > 0 {
		err := proto.Unmarshal(message, rpcMsg)

		if err != nil {
			fmt.Println("Invalid Message")
			return errors.New("Invlid Message: " + err.Error())
		}

		if rpcMsg.Type == uint32(ReqType) {
			cmd := RpcCmd(rpcMsg.Cmd)
			if processor, ok := Processors[cmd]; ok {
				rpcContext := NewContext(c)
				processor(ctx, rpcContext, rpcMsg)
			}
		} else {
			c.PromiseProcess(rpcMsg)
		}

	}

	return nil
}
