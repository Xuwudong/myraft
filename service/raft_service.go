package service

import (
	"_9932xt/myraft/gen-go/raft"
	"_9932xt/myraft/rpc"
	"_9932xt/myraft/state"
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

func AppendLogEntries(ctx context.Context, req *raft.AppendEntriesReq) error {
	var ch = make(chan int)
	var needCount = uint32(0)
	var once sync.Once
	for id, addr := range state.GetServerState().Conf.InnerAddrMap {
		if addr != state.GetServerState().Net.ServerAddr {
			tempAddr := addr
			tempReq := req
			tempId := id
			go func(id int, req *raft.AppendEntriesReq) {
				defer func() {
					atomic.AddUint32(&needCount, 1)
					if needCount >= state.GetMaxNum()-1 {
						once.Do(func() {
							ch <- 1
						})
					}
				}()
				log.Printf("id:%d", id)
				nextIndex, ok := state.GetServerState().MasterVolatileState.NextIndexMap[id]
				if !ok {
					state.SetMasterVolatileState(id, int64(len(state.GetServerState().PersistentState.Logs)), 0)
					nextIndex = state.GetServerState().MasterVolatileState.NextIndexMap[id]
				}
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
					resp, err := rpc.AppendEntriesByServer(id, tempAddr, req, true)
					if err != nil {
						// 超时等错误，重试
						log.Printf("Append Entries err:%v\n", err)
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
						log.Printf("递减重试，nextIndex:%d\n", nextIndex)
						time.Sleep(100 * time.Millisecond)
					}
				}
			}(tempId, tempReq)
		}
	}
	<-ch
	// 大多数复制成功
	log.Println("the logs has been copied by most of server")

	return nil
}
