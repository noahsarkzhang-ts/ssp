package network

import (
	"context"
	"io"
	"log"

	"github.com/ssp/util"
)

func FlowForward(ctx context.Context, client *Channel, target *RemoteConn) {

	traceId, _ := ctx.Value("traceId").(string)

	ch2connForward := func(src *Channel, dest *RemoteConn) {

		defer util.Trace(traceId, "ch2connForward")()
		defer src.Close()
		defer dest.Close()

		log.Printf("%s,Start ch2conn: %s\n", traceId, client)

		var written int64
		var err error
		if src.Available() && dest.Available() {
			written, err = io.Copy(dest.Target, src)
		}

		log.Printf("%s,Close ch2conn: %s,written:%d\n", traceId, client, written)
		if err != nil {
			log.Printf("%s,Close ch2conn: %s,case:%s\n", traceId, client, err.Error())
		}
	}

	conn2chForward := func(src *RemoteConn, dest *Channel) {
		defer util.Trace(traceId, "conn2chForward")()
		defer src.Close()
		defer dest.Close()

		log.Printf("%s,Start conn2ch: %s\n", traceId, client)

		var written int64
		var err error

		if src.Available() && dest.Available() {
			written, err = io.Copy(dest, src.Target)
		}

		log.Printf("%s,Close conn2ch: %s,written:%d\n", traceId, client, written)
		if err != nil {
			log.Printf("%s,Close conn2ch: %s,case:%s\n", traceId, client, err.Error())
		}

	}

	go ch2connForward(client, target)
	go conn2chForward(target, client)
}
