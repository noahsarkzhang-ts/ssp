package network

import (
	"fmt"
	"io"
	"net"
)

func FlowForward(client *Channel, target net.Conn) {
	ch2connForward := func(src *Channel, dest net.Conn) {
		defer src.Close()
		defer dest.Close()

		fmt.Printf("start ch2conn: %s\n", client)

		written, err := io.Copy(dest, src)

		fmt.Printf("close ch2conn: %s,written:%d\n", client, written)
		if err != nil {
			fmt.Printf("close ch2conn: %s,case:%s\n", client, err.Error())
		}
	}

	conn2chForward := func(src net.Conn, dest *Channel) {
		defer src.Close()
		defer dest.Close()

		fmt.Printf("start conn2ch: %s\n", client)
		written, err := io.Copy(dest, src)

		fmt.Printf("close conn2ch: %s,written:%d\n", client, written)
		if err != nil {
			fmt.Printf("close conn2ch: %s,case:%s\n", client, err.Error())
		}

	}

	go ch2connForward(client, target)
	go conn2chForward(target, client)
}
