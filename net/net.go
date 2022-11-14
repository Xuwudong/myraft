package net

import "_9932xt/myraft/gen-go/raft"

type Net struct {
	// 服务端间通信addr
	ServerAddr string
	// 客户端-服务端通信addr
	ClientAddr string

	PeerServerMap map[int]*PeerServer
}

type PeerServer struct {
	ServerAddr string
	Client     *raft.RaftServerClient
}
