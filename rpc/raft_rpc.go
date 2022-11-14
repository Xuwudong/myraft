package rpc

import (
	"_9932xt/myraft/gen-go/raft"
	"context"
	"fmt"
)

func AppendEntries(client *raft.RaftServerClient, ctx context.Context, req *raft.AppendEntriesReq) (*raft.AppendEntriesResp, error) {
	res, err := client.AppendEntries(ctx, req)
	if err != nil {
		fmt.Println("Error during operation:", err)
	}
	return res, err
}

func RequestVote(client *raft.RaftServerClient, ctx context.Context, req *raft.RequestVoteReq) (*raft.RequestVoteResp, error) {
	res, err := client.RequestVote(ctx, req)
	if err != nil {
		fmt.Println("Error during operation:", err)
	}
	return res, err
}
