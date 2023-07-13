package network

import (
	"log"
	"net"
	"sync"
)

type remoteConnFlag uint8

const (
	openFlag  remoteConnFlag = 1
	closeFlag remoteConnFlag = 2
)

type RemoteConn struct {
	sync.Mutex
	// 远程连接对象
	Target net.Conn

	flag remoteConnFlag

	TraceId string
}

func NewRemoteConn(target net.Conn) *RemoteConn {
	remoteConn := &RemoteConn{}

	remoteConn.Target = target
	remoteConn.flag = openFlag

	return remoteConn
}

func (r *RemoteConn) Close() error {

	r.Lock()
	defer r.Unlock()

	if r.flag == openFlag {

		r.flag = closeFlag
		err := r.Target.Close()

		log.Printf("%s,Close remote conn: %s:%s\n", r.TraceId, r.Target.LocalAddr(), r.Target.RemoteAddr())

		return err
	}

	return nil
}

func (r *RemoteConn) Available() bool {
	r.Lock()
	defer r.Unlock()

	return r.flag == openFlag
}
