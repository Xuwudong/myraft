package main

import (
	"context"
	"github.com/Xuwudong/myraft/gen-go/raft"
	"log"
)

var leader = "localhost:9090"

var members = []*raft.Member{{
	MemberID:   1,
	ClientAddr: "localhost:9090",
	ServerAddr: "localhost:8080",
},
}

func DoCommand(ctx context.Context, req *raft.DoCommandReq) (*raft.DoCommandResp, error) {
	var (
		resp   *raft.DoCommandResp
		client *Client
		err    error
	)
	if leader == "" {
		client, err = GetClient()
	} else {
		client, err = GetClientByServer(leader)
	}
	for {
		if err != nil {
			log.Printf("get client error:%v", err)
			return nil, err
		}
		resp, err = client.Client.DoCommand(ctx, req)
		if err != nil {
			log.Printf("do command err:%v", err)
			return nil, err
		}
		err := Return(client)
		if err != nil {
			log.Printf("return err:%v", err)
			return nil, err
		}
		if req.Command.Opt == raft.Opt_Write && resp.Leader != nil {
			//log.Printf("write req turn to the leader:%s", resp.GetLeader())
			leader = resp.GetLeader()
			client, err = GetClientByServer(leader)
		} else {
			break
		}
	}
	return resp, nil
}
