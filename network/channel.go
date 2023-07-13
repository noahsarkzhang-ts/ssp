package network

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"log"
)

type channelFlag uint8

const (
	channelOpenFlag  channelFlag = 1
	channelCloseFlag channelFlag = 2
)

type Channel struct {
	sync.Mutex

	// 通道id
	Id uint32

	// 读缓存
	ReadBuff chan []byte

	// Connection
	UnderlyingConn *Connection

	// 状态
	flag channelFlag

	// traceId
	TraceId string
}

func NewChannel(id uint32, conn *Connection) *Channel {
	channel := new(Channel)

	channel.ReadBuff = make(chan []byte, 1024)
	channel.UnderlyingConn = conn
	channel.Id = id
	channel.flag = channelOpenFlag

	return channel
}

func (c *Channel) Write(p []byte) (n int, err error) {
	wrLen := len(p)

	flowMsg := BuildMsgOfFlow(p, c.Id)
	SendMessge(context.TODO(), c.UnderlyingConn, flowMsg)

	return wrLen, nil
}

func (c *Channel) Read(p []byte) (n int, err error) {

	for data := range c.ReadBuff {
		cLen := copy(p, data)
		return cLen, nil
	}

	return -1, errors.New("Channel Close.")
}

func (c *Channel) Close() error {

	c.Lock()
	defer c.Unlock()

	if c.flag == channelCloseFlag {
		return nil
	}

	c.flag = channelCloseFlag
	close(c.ReadBuff)

	c.UnderlyingConn.RemoveChannel(c.Id)

	log.Printf("%s,Close channel %s \n", c.TraceId, c.String())

	return nil
}

func (c *Channel) AppendReadBuff(data []byte) {
	c.Lock()
	defer c.Unlock()

	if c.flag != channelCloseFlag {
		c.ReadBuff <- data
	}

}

func (c *Channel) String() string {
	return fmt.Sprintf("%s-%s:%d", c.UnderlyingConn.conn.RemoteAddr(), c.UnderlyingConn.conn.LocalAddr(), c.Id)
}

func (c *Channel) Available() bool {
	return c.flag == channelOpenFlag
}
