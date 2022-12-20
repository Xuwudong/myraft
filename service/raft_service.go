package service

import (
	"context"
	"fmt"
	"github.com/Xuwudong/myraft/conf"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/rpc"
	"github.com/Xuwudong/myraft/state"
)

func AppendLogEntriesToMost(ctx context.Context, req *raft.AppendEntriesReq) error {
	if state.GetServerState().MemberConf.State == conf.COld {
		AppendToMostServers(ctx, state.GetServerState().VolatileState.PeerServers, req)
	} else if state.GetServerState().MemberConf.State == conf.COldNew {
		AppendToMostServers(ctx, state.GetServerState().VolatileState.PeerServers, req)
		AppendToMostServers(ctx, state.GetServerState().VolatileState.NewPeerServers, req)
	} else {
		return fmt.Errorf("memberconf state error")
	}
	// 大多数复制成功
	logger.WithContext(ctx).Println("the logs has been copied by most of server")
	// update master commitIndex
	// 假设存在 N 满足N > commitIndex，使得大多数的 matchIndex[i] ≥ N 以及log[N].term == currentTerm 成立，则令 commitIndex = N（5.3 和 5.4 节）
	N := int64(len(state.GetServerState().PersistentState.Logs)) - 1
	for ; N > state.GetServerState().VolatileState.CommitIndex; N-- {
		if state.GetServerState().PersistentState.Logs[N].Term == state.GetServerState().PersistentState.CurrentTerm {
			atomic.StoreInt64(&state.GetServerState().VolatileState.CommitIndex, N)
			break
		}
	}
	return nil
}

func AppendToMostServers(ctx context.Context, peerServers []string, req *raft.AppendEntriesReq) {
	if len(peerServers) == 0 {
		return
	}
	var ch = make(chan int)
	var needCount = uint32(0)
	var once sync.Once
	for id, addr := range peerServers {
		tempAddr := addr
		tempReq := req
		tempId := id
		go func(id int, req *raft.AppendEntriesReq) {
			defer func() {
				atomic.AddUint32(&needCount, 1)
				if needCount+1 >= state.GetMaxNum(len(peerServers)) {
					once.Do(func() {
						ch <- 1
					})
				}
			}()
			value, ok := state.GetServerState().MasterVolatileState.NextIndexMap.Load(id)
			if !ok {
				state.SetMasterVolatileState(id, int64(len(state.GetServerState().PersistentState.Logs)), 0)
				value, _ = state.GetServerState().MasterVolatileState.NextIndexMap.Load(id)
			}
			nextIndex, _ := value.(int64)
			for {
				startIndex := nextIndex
				if startIndex < 0 {
					startIndex = 0
				}
				if startIndex >= int64(len(state.GetServerState().PersistentState.Logs)) {
					startIndex = int64(len(state.GetServerState().PersistentState.Logs)) - 1
				}
				req.Entries = state.GetServerState().PersistentState.Logs[startIndex:len(state.GetServerState().PersistentState.Logs)]
				req.PreLogIndex = nextIndex - 1
				if req.PreLogIndex >= 0 {
					req.PreLogTerm = state.GetServerState().PersistentState.Logs[req.PreLogIndex].Term
				}
				resp, err := rpc.AppendEntriesByServer(ctx, id, tempAddr, req, true)
				if err != nil {
					// 超时等错误，重试
					logger.WithContext(ctx).Errorf("Append Entries err:%v\n", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}
				if resp.Term > state.GetServerState().PersistentState.CurrentTerm {
					state.ToFollower(resp.Term)
				}
				if resp.Succuess {
					nextIndex = int64(len(state.GetServerState().PersistentState.Logs))
					matchIndex := nextIndex - 1
					state.SetMasterVolatileState(id, nextIndex, matchIndex)
					break
				} else {
					// 没有匹配到,重试
					nextIndex--
					logger.WithContext(ctx).Warnf("递减重试，nextIndex:%d\n", nextIndex)
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(tempId, tempReq)
	}
	<-ch
}
