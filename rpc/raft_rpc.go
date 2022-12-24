package rpc

import (
	"context"
	"time"

	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/pool"
)

func AppendEntries(client *raft.RaftServerClient, ctx context.Context, req *raft.AppendEntriesReq) (*raft.AppendEntriesResp, error) {
	res, err := client.AppendEntries(ctx, req)
	if err != nil {
		logger.WithContext(ctx).Errorf("Error during AppendEntries:%v", err)
	}
	return res, err
}

func RequestVote(client *raft.RaftServerClient, ctx context.Context, req *raft.RequestVoteReq) (*raft.RequestVoteResp, error) {
	res, err := client.RequestVote(ctx, req)
	if err != nil {
		logger.WithContext(ctx).Errorf("Error during RequestVote:%v", err)
	}
	return res, err
}

func AppendEntriesByServer(ctx context.Context, serverId int, addr string, req *raft.AppendEntriesReq, ensureSuccess bool) (*raft.AppendEntriesResp, error) {
	var client *pool.Client
	var err error
	for {
		client, err = pool.GetClientByServer(addr)
		if err != nil {
			logger.WithContext(ctx).Errorf("getClientByServer err:%v", err)
			if ensureSuccess {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return nil, err
		}
		break
	}

	//logger.WithContext(ctx).Infof("append log req:%v,id:%d", req, serverId)
	resp, err := AppendEntries(client.Client, ctx, req)
	defer func(client *pool.Client) {
		err := pool.Recycle(client)
		if err != nil {
			logger.WithContext(ctx).Errorf("Recycle error:%v", err)
		}
	}(client)
	if err != nil {
		logger.WithContext(ctx).Errorf("Append Entries err:%v", err)
		defer func() {
			err2 := pool.Recycle(client)
			if err2 != nil {
				logger.WithContext(ctx).Errorf("recycle err:%v", err2)
			}
		}()
		return resp, err
	}

	return resp, nil
}
