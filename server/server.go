package server

import (
	"fmt"
	"log"
	"net"

	"github.com/ssp/network"
)

type ServerFlag int

const (
	Init    ServerFlag = 0
	Ready   ServerFlag = 1
	UnReady ServerFlag = -1
)

type Server struct {
	Flag ServerFlag
	Port int
}

func New(port int) *Server {
	return &Server{Init, port}
}

func (s *Server) Start() {

	addr := fmt.Sprintf("%s:%d", "localhost", s.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("Server %s start fail...\n", addr)
		panic(err)
	}

	log.Printf("Server %s start successfuly...\n", addr)

	go s.Accept(listener)

}

func (s *Server) Accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		log.Printf("New conn:%s \n", conn.RemoteAddr())

		connection := network.NewConnection(conn)

		go connection.Read()
		go connection.Write()
		go connection.Timeout()
	}
}
