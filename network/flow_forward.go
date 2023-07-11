package network

import (
	"io"
	"log"
)

func FlowForward(client *Channel, target *RemoteConn) {

	ch2connForward := func(src *Channel, dest *RemoteConn) {
		defer src.Close()
		defer dest.Close()

		log.Printf("start ch2conn: %s\n", client)

		var written int64
		var err error
		if src.Available() && dest.Available() {
			written, err = io.Copy(dest.Target, src)
		}

		log.Printf("close ch2conn: %s,written:%d\n", client, written)
		if err != nil {
			log.Printf("close ch2conn: %s,case:%s\n", client, err.Error())
		}
	}

	conn2chForward := func(src *RemoteConn, dest *Channel) {
		defer src.Close()
		defer dest.Close()

		log.Printf("start conn2ch: %s\n", client)

		var written int64
		var err error

		if src.Available() && dest.Available() {
			written, err = io.Copy(dest, src.Target)
		}

		log.Printf("close conn2ch: %s,written:%d\n", client, written)
		if err != nil {
			log.Printf("close conn2ch: %s,case:%s\n", client, err.Error())
		}

	}

	go ch2connForward(client, target)
	go conn2chForward(target, client)
}
