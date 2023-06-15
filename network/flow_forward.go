package network

import (
	"io"
	"net"
)

func FlowForward(client *Channel, target net.Conn) {
	ch2connForward := func(src *Channel, dest net.Conn) {
		defer src.Close()
		defer dest.Close()

		io.Copy(dest, src)
	}

	conn2chForward := func(src net.Conn, dest *Channel) {
		defer src.Close()
		defer dest.Close()

		io.Copy(dest, src)
	}

	go ch2connForward(client, target)
	go conn2chForward(target, client)
}
