package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ssp/client"
	"github.com/ssp/util"
)

func main() {
	serverAddr := "localhost:9090"

	proxy := client.New(serverAddr)

	proxy.Connect()

	proxy.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	gid := util.GetGID()

	for {
		s := <-c

		log.Printf("gid:%d,Receive a signal!!!\n", gid)

		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Printf("gid:%d,Proxy exist!!!\n", gid)
			return
		case syscall.SIGHUP:
		default:
			return
		}

	}

}
