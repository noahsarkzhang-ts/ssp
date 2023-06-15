package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ssp/server"
)

func main() {
	port := 9090

	server := server.New(port)

	server.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	for {
		s := <-c

		fmt.Println("Receive a signal!!!")

		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			fmt.Println("Server exist!!!")
			return
		case syscall.SIGHUP:
		default:
			return
		}

	}
}
