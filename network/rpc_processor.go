package network

var Processors map[RpcCmd]RpcProcessor = make(map[RpcCmd]RpcProcessor)

func init() {
	Processors[LoginCmd] = Login
	Processors[BuildChannelCmd] = BuildNewChannel
}
