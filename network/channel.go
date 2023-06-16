package network

import "fmt"

type Channel struct {
	// 通道id
	Id uint32

	// 读缓存
	ReadBuff chan []byte

	// Connection
	UnderlyingConn *Connection
}

func NewChannel(id uint32, conn *Connection) *Channel {
	channel := new(Channel)

	channel.ReadBuff = make(chan []byte, 1024)
	channel.UnderlyingConn = conn
	channel.Id = id

	return channel
}

func (c *Channel) Write(p []byte) (n int, err error) {
	wrLen := len(p)

	flowMsg := BuildMsgOfFlow(p, c.Id)
	SendMessge(c.UnderlyingConn, flowMsg)

	return wrLen, nil
}

func (c *Channel) Read(p []byte) (n int, err error) {

	for data := range c.ReadBuff {
		cLen := copy(p, data)
		return cLen, nil
	}

	return 0, nil
}

func (c *Channel) Close() error {

	//close(c.ReadBuff)

	return nil
}

func (c *Channel) AppendReadBuff(data []byte) {
	c.ReadBuff <- data
}

func (c *Channel) String() string {
	return fmt.Sprintf("%s-%s:%d", c.UnderlyingConn.conn.RemoteAddr(), c.UnderlyingConn.conn.LocalAddr(), c.Id)
}
