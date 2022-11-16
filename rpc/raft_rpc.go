package rpc

import (
	"_9932xt/myraft/errno"
	"_9932xt/myraft/gen-go/raft"
	"_9932xt/myraft/net"
	"context"
	"log"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
)

func AppendEntries(client *raft.RaftServerClient, ctx context.Context, req *raft.AppendEntriesReq) (*raft.AppendEntriesResp, error) {
	res, err := client.AppendEntries(ctx, req)
	if err != nil {
		log.Println("Error during AppendEntries:", err)
	}
	return res, err
}

func RequestVote(client *raft.RaftServerClient, ctx context.Context, req *raft.RequestVoteReq) (*raft.RequestVoteResp, error) {
	res, err := client.RequestVote(ctx, req)
	if err != nil {
		log.Println("Error during RequestVote:", err)
	}
	return res, err
}

func AppendEntriesByServer(serverId int, psTemp *net.PeerServer, req *raft.AppendEntriesReq, newClientRetry bool) (*raft.AppendEntriesResp, error) {
	var (
		client    *raft.RaftServerClient
		err       error
		transport thrift.TTransport
	)
	for {
		client, transport, err = NewClient(net.TransportFactory, net.ProtocolFactory, psTemp.ServerAddr, net.Secure, net.Cfg)
		if err == nil {
			break
		}
		if !newClientRetry {
			return nil, errno.NewNewClientErr(err.Error())
		}
		time.Sleep(1 * time.Millisecond)
		log.Printf("error new client: %v", err)
	}
	log.Printf("append log req:%v,id:%d", req, serverId)
	defer func(transport thrift.TTransport) {
		err := transport.Close()
		if err != nil {
			log.Printf("transport close err:%v", err)
		}
	}(transport)
	resp, err := AppendEntries(client, context.Background(), req)
	if err != nil {
		log.Printf("Append Entries err:%v\n", err)
		return resp, err
	}
	return resp, nil
}
