package network

import (
	"log"

	"github.com/ssp/msg"
	"google.golang.org/protobuf/proto"
)

type Context struct {

	// 连接对象
	conn *Connection
}

func NewContext(conn *Connection) *Context {
	context := &Context{}

	context.conn = conn

	return context
}

func (c *Context) SendMessge(message *msg.Msg) {
	bMsg, err := proto.Marshal(message)

	if err != nil {
		return
	}

	data, err := msg.Encode(bMsg)
	if err != nil {
		log.Printf("send %d \n", len(data))
		return
	}

	c.conn.WriteBytes(data)
}
