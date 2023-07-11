package network

import (
	"log"
	"net"
	"time"

	"github.com/ssp/msg"
	"google.golang.org/protobuf/proto"
)

func RpcInvoker(conn *Connection, message *msg.RpcMsg, timeout time.Duration, callbck RpcCallback) *RpcPromise {

	rpcMsg := BuildMsgOfRpc(message)
	log.Printf("send message: %v \n", rpcMsg)

	SendMessge(conn, rpcMsg)

	promise := NewRpcPromise(timeout, callbck)
	conn.RegPromise(message.Id, promise)

	return promise

}

func SendMessge(conn *Connection, message *msg.Msg) {
	bMsg, err := proto.Marshal(message)

	if err != nil {
		return
	}

	data, err := msg.Encode(bMsg)
	if err != nil {
		log.Printf("send %d \n", len(data))
		return
	}

	conn.WriteBytes(data)
}

func BuildNewChannel(context *Context, message *msg.RpcMsg) {
	channelReq := &msg.NewChannelReq{}

	err := proto.Unmarshal(message.Data, channelReq)
	if err != nil {
		log.Println("Invlid new channel request!")
		return
	}

	log.Printf("Receive a new channel request:%+v \n", channelReq)

	// TODO
	channel := context.conn.ApplyChannel()

	// 建立 TCP 连接
	// TODO

	var rpcMsg *msg.RpcMsg
	var resMsg *msg.Msg
	destAddrPort := channelReq.Addr
	dest, err := net.Dial("tcp", destAddrPort)

	if err != nil {

		log.Printf("net Dial error:%s \n", err.Error())

		rpcMsg = BuildNewChannelRes(message, channel, -1, err.Error())
		resMsg = BuildMsgOfRpc(rpcMsg)

		context.SendMessge(resMsg)

		return
	}

	// 转发
	target := NewRemoteConn(dest)
	FlowForward(channel, target)

	rpcMsg = BuildNewChannelRes(message, channel, 1, "success")
	resMsg = BuildMsgOfRpc(rpcMsg)
	context.SendMessge(resMsg)

}

func Login(context *Context, message *msg.RpcMsg) {
	log.Printf("receive a login request:%+v\n", message)

	res := BuildCommonRes(message)
	resMsg := BuildMsgOfRpc(res)

	context.SendMessge(resMsg)

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

func BuildNewChannelReq(conn *Connection, addr string) *msg.RpcMsg {
	request := BuildRequestHeader(conn, BuildChannelCmd)

	channelReq := &msg.NewChannelReq{}
	channelReq.Addr = addr

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

func BuildNewChannelRes(req *msg.RpcMsg, ch *Channel, code int32, resString string) *msg.RpcMsg {
	res := BuildResponseHeader(req)

	channelRes := &msg.NewChannelRes{}
	channelRes.Code = code
	channelRes.Msg = resString
	channelRes.ChannelId = ch.Id

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
