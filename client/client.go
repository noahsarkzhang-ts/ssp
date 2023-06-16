package client

import (
	"errors"
	"log"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/ssp/msg"
	"github.com/ssp/network"
)

type ClientFlag int

const (
	Init        ClientFlag = 0
	Connected   ClientFlag = 1
	UnConnected ClientFlag = 2
	Ready       ClientFlag = 3
)

type Client struct {
	Flag       ClientFlag
	ServerAddr string
	RemoteConn *network.Connection
	Proxy      *Socks5Proxy
}

func New(ServerAddr string) *Client {
	return &Client{Flag: Init, ServerAddr: ServerAddr}
}

func (c *Client) Connect() {

	conn, err := net.Dial("tcp", c.ServerAddr)
	if err != nil {
		log.Printf("Connect server:%s fail...\n", c.ServerAddr)
		panic(err)
	}

	log.Printf("Connect server:%s success...\n", c.ServerAddr)

	connection := network.NewConnection(conn)
	c.RemoteConn = connection

	go connection.Read()
	go connection.Write()

	c.login()

}

func (c *Client) login() {
	message := network.BuildLoginReq(c.RemoteConn)
	log.Printf("send rpc message: %v \n", message)

	promise := network.RpcInvoker(c.RemoteConn, message, 5*time.Second, nil)

	res, ok := promise.Get()

	if !ok {
		log.Println("login request fail")
		c.Flag = UnConnected
		return
	}

	data := res.Data
	commonRes := &msg.CommonRes{}

	err := proto.Unmarshal(data, commonRes)
	if err != nil {
		log.Println("Invlid login res!!!")
		c.Flag = UnConnected
		return
	}

	// 成功
	if commonRes.Code == 1 {
		c.Flag = Ready
	}

}

func (c *Client) BuildNewChannel(addr string) (*network.Channel, error) {

	channelMessage := network.BuildNewChannelReq(c.RemoteConn, addr)
	log.Printf("send new channel message: %v \n", channelMessage)

	channelPromise := network.RpcInvoker(c.RemoteConn, channelMessage, 5*time.Second, nil)
	res, ok := channelPromise.Get()

	if !ok {
		log.Println("new channel request fail")

		return nil, errors.New("New channel fail")

	}

	data := res.Data
	channelRes := &msg.NewChannelRes{}

	err := proto.Unmarshal(data, channelRes)
	if err != nil {
		log.Println("Invlid channel res!!!")

		return nil, errors.New("Invlid channel res")
	}

	// 成功
	if channelRes.Code == 1 {

		channelId := channelRes.ChannelId
		channel := network.NewChannel(channelId, c.RemoteConn)

		c.RemoteConn.RegChannel(channelId, channel)

		return channel, nil

	}

	return nil, errors.New("Build channel fail")

}

func (c *Client) Start() {

	addr := ":1080"

	proxy := NewSocks5Proxy(addr, c)
	c.Proxy = proxy

	proxy.Start()
}
