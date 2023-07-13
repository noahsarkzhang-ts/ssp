package network

import (
	"log"
	"time"

	"github.com/ssp/msg"
)

type RpcPromise struct {
	timer    *time.Timer
	result   chan *msg.RpcMsg
	callback RpcCallback
	traceId  string
}

func NewRpcPromise(traceId string, timeout time.Duration, callback RpcCallback) *RpcPromise {
	promise := &RpcPromise{}

	promise.timer = time.NewTimer(timeout)
	promise.result = make(chan *msg.RpcMsg)
	promise.callback = callback
	promise.traceId = traceId

	return promise
}

func (p *RpcPromise) Get() (*msg.RpcMsg, bool) {
	var res *msg.RpcMsg

	select {
	case res := <-p.result:
		p.timer.Stop()

		if p.callback != nil {
			p.callback(res)
		}

		return res, true
	case <-p.timer.C: //超时
		log.Printf("%s,RpcPromise timeout.\n", p.traceId)
		return res, false
	}
}

func (p *RpcPromise) Set(res *msg.RpcMsg) bool {
	p.result <- res

	if p.callback != nil {
		p.callback(res)
	}

	return true
}
