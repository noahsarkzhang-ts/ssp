package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ssp/client"
)

func main() {
	serverAddr := "localhost:9090"

	proxy := client.New(serverAddr)

	proxy.Connect()

	proxy.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	for {
		s := <-c

		fmt.Println("Receive a signal!!!")

		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			fmt.Println("Proxy exist!!!")
			return
		case syscall.SIGHUP:
		default:
			return
		}

	}

}
