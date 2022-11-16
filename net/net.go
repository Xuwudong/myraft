package net

import (
	"github.com/apache/thrift/lib/go/thrift"
)

type Net struct {
	// 服务端间通信addr
	ServerAddr string
	// 客户端-服务端通信addr
	ClientAddr string

	// 存活的对端服务列表
	PeerServerMap map[int]*PeerServer
}

type PeerServer struct {
	ServerAddr string
	//Client     *raft.RaftServerClient
}

var TransportFactory thrift.TTransportFactory
var ProtocolFactory thrift.TProtocolFactory
var Secure bool
var Cfg *thrift.TConfiguration
