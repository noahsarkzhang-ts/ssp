package client

import (
	"context"
	"errors"
	"log"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/ssp/msg"
	"github.com/ssp/network"
	"github.com/ssp/util"
)

type ClientFlag int

const (
	Init        ClientFlag = 0
	Connected   ClientFlag = 1
	UnConnected ClientFlag = 2
	Ready       ClientFlag = 3
	UnReady     ClientFlag = 4
)

type Client struct {
	Flag       ClientFlag
	ServerAddr string
	RemoteConn *network.Connection
	Proxy      *Socks5Proxy

	ticker time.Ticker
}

func New(ServerAddr string) *Client {
	return &Client{Flag: Init, ServerAddr: ServerAddr}
}

func (c *Client) Connect() bool {
	defer util.Trace("", "Client Connect")()

	conn, err := net.Dial("tcp", c.ServerAddr)
	if err != nil {
		log.Printf("Connect remote server:%s fail...\n", c.ServerAddr)
		c.Flag = UnConnected
		return false
	}

	log.Printf("Connect remote server:%s success...\n", c.ServerAddr)

	connection := network.NewConnection(conn)
	c.RemoteConn = connection

	go connection.Read()
	go connection.Write()
	go connection.PingPongAndTimeout()

	c.login()

	return c.Flag == Ready

}

func (c *Client) Reconnect() {
	c.ticker = *time.NewTicker(5 * time.Second)

	for {
		select {
		case <-c.ticker.C:
			if c.Flag != Ready || c.RemoteConn.Closed() {
				log.Printf("Connection was close,reconnect ......\n")
				c.Connect()
			}
		}
	}
}

func (c *Client) login() {

	defer util.Trace("", "Client Login")()

	message := network.BuildLoginReq(c.RemoteConn)
	log.Printf("send rpc message: %v \n", message)

	promise := network.RpcInvoker(context.TODO(), c.RemoteConn, message, 5*time.Second, nil)

	res, ok := promise.Get()

	if !ok {
		log.Println("login request fail")
		c.Flag = UnReady
		return
	}

	data := res.Data
	commonRes := &msg.CommonRes{}

	err := proto.Unmarshal(data, commonRes)
	if err != nil {
		log.Println("Invlid login res!!!")
		c.Flag = UnReady
		return
	}

	// 成功
	if commonRes.Code == 1 {
		c.Flag = Ready
	}

}

func (c *Client) BuildNewChannel(ctx context.Context, addr string) (*network.Channel, error) {

	if !c.Available() {
		log.Println("Client is not available!")

		return nil, errors.New("Client is not available!")
	}

	traceId, _ := ctx.Value("traceId").(string)

	defer util.Trace(traceId, "Client BuildNewChannel")()

	channelMessage := network.BuildNewChannelReq(c.RemoteConn, addr, traceId)
	log.Printf("%s,send new channel message: %v \n", traceId, channelMessage)

	channelPromise := network.RpcInvoker(ctx, c.RemoteConn, channelMessage, 5*time.Second, nil)
	res, ok := channelPromise.Get()

	if !ok {
		log.Printf("%s,new channel request fail.\n", traceId)

		return nil, errors.New("New channel fail")

	}

	data := res.Data
	channelRes := &msg.NewChannelRes{}

	err := proto.Unmarshal(data, channelRes)
	if err != nil {
		log.Printf("%s,Invlid channel res!!!\n", traceId)

		return nil, errors.New("Invlid channel res")
	}

	// 成功
	if channelRes.Code == 1 {

		channelId := channelRes.ChannelId
		channel := network.NewChannel(channelId, c.RemoteConn)
		channel.TraceId = traceId

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

func (c *Client) Available() bool {
	return c.Flag == Ready
}
