package net

import (
	"github.com/apache/thrift/lib/go/thrift"
)

type Net struct {
	// 服务端间通信addr
	ServerAddr string
	// 客户端-服务端通信addr
	ClientAddr string
}

type PeerServer struct {
	ServerAddr string
	//Client     *raft.RaftServerClient
}

var TransportFactory thrift.TTransportFactory
var ProtocolFactory thrift.TProtocolFactory
var Secure bool
var Cfg *thrift.TConfiguration
