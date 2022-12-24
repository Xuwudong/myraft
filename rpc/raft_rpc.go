package rpc

import (
	"context"
	"time"

	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/pool"
)

func AppendEntries(ctx context.Context, server string, req *raft.AppendEntriesReq, ensureSuccess bool) (*raft.AppendEntriesResp, error) {
	var resp *raft.AppendEntriesResp
	var err error
	for {
		resp, err = pool.AppendEntries(ctx, server, req)
		if err != nil {
			logger.WithContext(ctx).Errorf("AppendEntries err:%v", err)
			if ensureSuccess {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return nil, err
		}
		break
	}
	return resp, err
}

func RequestVote(ctx context.Context, server string, req *raft.RequestVoteReq) (*raft.RequestVoteResp, error) {
	retry := 0
	var res *raft.RequestVoteResp
	var err error
	for {
		res, err = pool.RequestVote(ctx, server, req)
		if err == nil {
			break
		}
		logger.WithContext(ctx).Errorf("Error during RequestVote:%v", err)
		retry++
		if retry > 5 {
			break
		}
	}
	return res, err
}

func AppendEntriesByServer(ctx context.Context, serverId int, server string, req *raft.AppendEntriesReq, ensureSuccess bool) (*raft.AppendEntriesResp, error) {
	resp, err := AppendEntries(ctx, server, req, ensureSuccess)
	if err != nil {
		logger.WithContext(ctx).Errorf("Append Entries err:%v", err)
		return resp, err
	}
	return resp, nil
}
