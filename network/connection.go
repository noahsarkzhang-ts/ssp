package network

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"

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
}

func NewConnection(conn net.Conn) *Connection {

	connection := new(Connection)
	connection.conn = conn
	connection.channelIdGenerator = util.NewId(0)
	connection.requestIdGenerator = util.NewId(0)
	connection.writerBuff = make(chan []byte, 1024)

	connection.channels = map[uint32]*Channel{}
	connection.promises = map[uint32]*RpcPromise{}

	return connection
}

func (c *Connection) Read() {

	reader := bufio.NewReader(c.conn)
	//读消息
	for {
		data, err := msg.Decode(reader)
		m := &msg.Msg{}

		if err == nil && len(data) > 0 {
			err := proto.Unmarshal(data, m)
			if err != nil {
				log.Printf("Close connection:%s \n", c.conn.RemoteAddr())
				c.Close()

				break
			}
		}
		if err != nil {
			log.Printf("Close connection:%s \n", c.conn.RemoteAddr())
			c.Close()

			break
		}

		cmd := MsgCmd(m.Cmd)
		switch cmd {
		case RpcMsgCmd:
			go c.RpcProcess(m.Data)
		case FlowMsgCmd:
			// 写入channel
			c.Flow(m)
		case HeartbeatMsgCmd:
		default:

		}

	}

}

func (c *Connection) Close() {
	// 关闭 写缓存
	close(c.writerBuff)

	// 关闭通道
	for id, ch := range c.channels {

		log.Printf("Close channel: %d \n", id)
		ch.Close()
	}

	c.channels = nil
	c.promises = nil

}

func (c *Connection) Write() {

	for data := range c.writerBuff {
		c.conn.Write(data)
	}

}

func (c *Connection) WriteBytes(data []byte) error {

	c.writerBuff <- data

	return nil
}

func (c *Connection) ApplyChannel() *Channel {

	// 申请一个唯一的通道 id
	id := c.channelIdGenerator.IncrementAndGet()
	channel := NewChannel(id, c)

	c.channels[id] = channel

	return channel

}

func (c *Connection) RegChannel(channelId uint32, channel *Channel) bool {
	c.channels[channelId] = channel

	return true
}

func (c *Connection) RemoveChannel(channelId uint32) bool {
	delete(c.channels, channelId)

	return true
}

func (c *Connection) Flow(msg *msg.Msg) {
	id := msg.Id

	if channel, ok := c.channels[id]; ok {
		channel.AppendReadBuff(msg.Data)
	}
}

func (c *Connection) RegPromise(requestId uint32, promise *RpcPromise) bool {

	c.promises[requestId] = promise

	return true
}

func (c *Connection) RpcProcess(message []byte) error {
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
				context := NewContext(c)
				processor(context, rpcMsg)
			}
		} else {
			requestId := rpcMsg.Id
			if promise, ok := c.promises[requestId]; ok {
				promise.Set(rpcMsg)
			}
		}

	}

	return nil
}
