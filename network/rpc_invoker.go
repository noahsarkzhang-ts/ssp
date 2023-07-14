package network

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/ssp/msg"
	"github.com/ssp/util"
	"google.golang.org/protobuf/proto"
)

func RpcInvoker(ctx context.Context, conn *Connection, message *msg.RpcMsg, timeout time.Duration, callbck RpcCallback) *RpcPromise {

	traceId, _ := ctx.Value("traceId").(string)

	defer util.Trace(traceId, "Client RpcInvoker")()

	rpcMsg := BuildMsgOfRpc(message)
	log.Printf("%s,send message: %v \n", traceId, rpcMsg)

	SendMessge(ctx, conn, rpcMsg)

	promise := NewRpcPromise(traceId, timeout, callbck)
	conn.RegPromise(message.Id, promise)

	return promise

}

func SendMessge(ctx context.Context, conn *Connection, message *msg.Msg) {
	traceId, _ := ctx.Value("traceId").(string)

	bMsg, err := proto.Marshal(message)

	if err != nil {
		return
	}

	data, err := msg.Encode(bMsg)
	if err != nil {
		log.Printf("%s,send %d \n", traceId, len(data))
		return
	}

	conn.WriteBytes(data)
}

func BuildNewChannel(ctx context.Context, rpcContext *Context, message *msg.RpcMsg) {

	var rpcMsg *msg.RpcMsg
	var resMsg *msg.Msg

	channelReq := &msg.NewChannelReq{}

	err := proto.Unmarshal(message.Data, channelReq)
	if err != nil {

		log.Printf("Invlid new channel request!:%s \n", err.Error())

		rpcMsg = BuildNewChannelRes(message, 0, -1, err.Error())
		resMsg = BuildMsgOfRpc(rpcMsg)

		rpcContext.SendMessge(resMsg)

		return
	}

	traceId := channelReq.TraceId
	newCtx := context.WithValue(ctx, "traceId", traceId)
	defer util.Trace(traceId, "Server BuildNewChannel")()

	log.Printf("%s,Receive a new channel request:%+v \n", traceId, channelReq)

	// TODO
	channel := rpcContext.conn.ApplyChannel()
	channel.TraceId = traceId

	// 建立 TCP 连接
	// TODO
	destAddrPort := channelReq.Addr
	dest, err := net.Dial("tcp", destAddrPort)

	if err != nil {

		log.Printf("%s,Net Dial error:%s \n", traceId, err.Error())

		rpcMsg = BuildNewChannelRes(message, channel.Id, -1, err.Error())
		resMsg = BuildMsgOfRpc(rpcMsg)

		rpcContext.SendMessge(resMsg)

		return
	}

	// 转发
	target := NewRemoteConn(dest)
	target.TraceId = traceId
	FlowForward(newCtx, channel, target)

	rpcMsg = BuildNewChannelRes(message, channel.Id, 1, "success")
	resMsg = BuildMsgOfRpc(rpcMsg)
	rpcContext.SendMessge(resMsg)

}

func Login(ctx context.Context, rpcContext *Context, message *msg.RpcMsg) {
	defer util.Trace("Server Login", "")()

	log.Printf("receive a login request:%+v\n", message)

	res := BuildCommonRes(message)
	resMsg := BuildMsgOfRpc(res)

	rpcContext.SendMessge(resMsg)

}

func BuildLoginReq(conn *Connection) *msg.RpcMsg {
	request := BuildRequestHeader(conn, LoginCmd)

	loginReq := &msg.LoginReq{}
	loginReq.Name = "Allen"
	loginReq.Pwd = "Allen"

	bLoginReq, err := proto.Marshal(loginReq)
	if err != nil {
		panic(err)
	}

	request.Data = bLoginReq

	return request

}

func BuildNewChannelReq(conn *Connection, addr string, traceId string) *msg.RpcMsg {
	request := BuildRequestHeader(conn, BuildChannelCmd)

	channelReq := &msg.NewChannelReq{}
	channelReq.Addr = addr
	channelReq.TraceId = traceId

	bChannelReq, err := proto.Marshal(channelReq)
	if err != nil {
		panic(err)
	}

	request.Data = bChannelReq

	return request

}

func BuildMsgOfRpc(message *msg.RpcMsg) *msg.Msg {
	msg := &msg.Msg{}

	msg.Id = 0
	msg.Cmd = uint32(RpcMsgCmd)

	rpcMsg, err := proto.Marshal(message)
	if err != nil {
		log.Println("Invlid rpc message!")
		panic(err)
	}

	msg.Data = rpcMsg

	return msg
}

func BuildMsgOfFlow(data []byte, channelId uint32) *msg.Msg {
	msg := &msg.Msg{}

	msg.Id = channelId
	msg.Cmd = uint32(FlowMsgCmd)

	msg.Data = data

	return msg
}

func BuildMsgOfPing() *msg.Msg {
	msg := &msg.Msg{}

	msg.Id = 1
	msg.Cmd = uint32(PingMsgCmd)

	return msg
}

func BuildMsgOfPong() *msg.Msg {
	msg := &msg.Msg{}

	msg.Id = 1
	msg.Cmd = uint32(PongMsgCmd)

	return msg
}

func LoginReqCallback(conn *Connection, message *msg.RpcMsg) {
	res := &msg.CommonRes{}

	err := proto.Unmarshal(message.Data, res)
	if err != nil {
		log.Println("Invalid callback")
		return
	}

	log.Printf("Receive a login response: %v \n", res)
}

func BuildCommonRes(req *msg.RpcMsg) *msg.RpcMsg {
	res := BuildResponseHeader(req)

	commonRes := &msg.CommonRes{}
	commonRes.Code = 1
	commonRes.Msg = "success"

	bCommonRes, err := proto.Marshal(commonRes)
	if err != nil {
		log.Println("Invlid Message")
		return nil
	}

	res.Data = bCommonRes

	return res

}

func BuildNewChannelRes(req *msg.RpcMsg, channelId uint32, code int32, resString string) *msg.RpcMsg {
	res := BuildResponseHeader(req)

	channelRes := &msg.NewChannelRes{}
	channelRes.Code = code
	channelRes.Msg = resString
	channelRes.ChannelId = channelId

	bChannelRes, err := proto.Marshal(channelRes)
	if err != nil {
		log.Println("Invlid Message")
		return nil
	}

	res.Data = bChannelRes

	return res

}

func BuildResponseHeader(req *msg.RpcMsg) *msg.RpcMsg {
	res := &msg.RpcMsg{}
	res.Type = uint32(ResType)
	res.Cmd = uint32(req.Cmd)
	res.Id = req.Id

	return res
}

func BuildRequestHeader(conn *Connection, cmd RpcCmd) *msg.RpcMsg {
	request := &msg.RpcMsg{}
	request.Type = uint32(ReqType)
	request.Cmd = uint32(cmd)

	request.Id = conn.requestIdGenerator.IncrementAndGet()

	return request
}
