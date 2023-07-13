package network

import (
	"context"

	"github.com/ssp/msg"
)

type RpcCmd uint32

const (
	LoginCmd        RpcCmd = 11
	BuildChannelCmd RpcCmd = 12
)

type RpcMsgType uint32

const (
	ReqType RpcMsgType = 1
	ResType RpcMsgType = 2
)

type RequestId uint32

type RpcProcessor func(ctx context.Context, rpcContext *Context, message *msg.RpcMsg)
type RpcCallback func(message *msg.RpcMsg)
