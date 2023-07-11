package client

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/ssp/network"
)

type Socks5Proxy struct {
	Addr           string
	Proxy          net.Listener
	RemoteEndpoint *Client
}

func NewSocks5Proxy(addr string, client *Client) *Socks5Proxy {
	return &Socks5Proxy{Addr: addr, RemoteEndpoint: client}
}

func (p *Socks5Proxy) Start() {
	server, err := net.Listen("tcp", p.Addr)
	if err != nil {
		log.Printf("Listen failed: %v\n", err)
		return
	}

	p.Proxy = server
	go p.Accept()

	log.Println("Proxy Listen successfully:")
}

func (p *Socks5Proxy) Accept() {
	for {
		src, err := p.Proxy.Accept()
		if err != nil {
			log.Printf("Socks5 proxy accept failed: %+v \n", err)
			continue
		}
		go p.Process(src)
	}
}

func (p *Socks5Proxy) Process(src net.Conn) {

	log.Printf("New conn:%s \n", src.RemoteAddr())

	if err := p.Socks5Auth(src); err != nil {
		log.Println("auth error:", err)
		src.Close()
		return
	}

	channel, err := p.Socks5Connect(src)
	if err != nil {
		log.Println("connect error:", err)
		src.Close()
		return
	}

	target := network.NewRemoteConn(src)
	network.FlowForward(channel, target)
}

func (p *Socks5Proxy) Socks5Auth(src net.Conn) (err error) {
	buf := make([]byte, 256)

	// 读取 VER 和 NMETHODS
	n, err := io.ReadFull(src, buf[:2])
	if n != 2 {
		return errors.New("reading header: " + err.Error())
	}

	ver, nMethods := int(buf[0]), int(buf[1])
	if ver != 5 {
		return errors.New("invalid version")
	}

	// 读取 METHODS 列表
	n, err = io.ReadFull(src, buf[:nMethods])
	if n != nMethods {
		return errors.New("reading methods: " + err.Error())
	}

	//无需认证
	n, err = src.Write([]byte{0x05, 0x00})
	if n != 2 || err != nil {
		return errors.New("write rsp: " + err.Error())
	}

	return nil
}

func (p *Socks5Proxy) Socks5Connect(src net.Conn) (*network.Channel, error) {
	buf := make([]byte, 256)

	n, err := io.ReadFull(src, buf[:4])
	if n != 4 {
		return nil, errors.New("read header: " + err.Error())
	}

	ver, cmd, _, atyp := buf[0], buf[1], buf[2], buf[3]
	if ver != 5 || cmd != 1 {
		return nil, errors.New("invalid ver/cmd")
	}

	addr := ""
	switch atyp {
	case 1:
		n, err = io.ReadFull(src, buf[:4])
		if n != 4 {
			return nil, errors.New("invalid IPv4: " + err.Error())
		}
		addr = fmt.Sprintf("%d.%d.%d.%d", buf[0], buf[1], buf[2], buf[3])

	case 3:
		n, err = io.ReadFull(src, buf[:1])
		if n != 1 {
			return nil, errors.New("invalid hostname: " + err.Error())
		}
		addrLen := int(buf[0])

		n, err = io.ReadFull(src, buf[:addrLen])
		if n != addrLen {
			return nil, errors.New("invalid hostname: " + err.Error())
		}
		addr = string(buf[:addrLen])

	case 4:
		return nil, errors.New("IPv6: no supported yet")

	default:
		return nil, errors.New("invalid atyp")
	}

	n, err = io.ReadFull(src, buf[:2])
	if n != 2 {
		return nil, errors.New("read port: " + err.Error())
	}
	port := binary.BigEndian.Uint16(buf[:2])

	destAddrPort := fmt.Sprintf("%s:%d", addr, port)

	// dest, err := net.Dial("tcp", destAddrPort)
	// 建立远程通道
	log.Printf("Connect %s\n", destAddrPort)
	dest, err := p.RemoteEndpoint.BuildNewChannel(destAddrPort)

	if err != nil {
		log.Printf("Connect %s failed\n", destAddrPort)
		src.Write([]byte{0x05, 0x04, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return nil, errors.New("dial dst: " + err.Error())
	}

	n, err = src.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err != nil {
		dest.Close()
		return nil, errors.New("write rsp: " + err.Error())
	}

	return dest, nil
}
