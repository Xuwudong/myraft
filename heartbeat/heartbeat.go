package heartbeat

import (
	"context"
	"fmt"
	"github.com/Xuwudong/myraft/conf"
	_ "net/http/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/rpc"
	"github.com/Xuwudong/myraft/state"
)

func Run() {
	for {
		if state.GetServerState().Role == raft.Role_Leader {
			if state.GetServerState().MemberConf.State == conf.COld {
				Heartbeat(state.GetServerState().VolatileState.PeerServers)
			} else if state.GetServerState().MemberConf.State == conf.COldNew {
				Heartbeat(state.GetServerState().VolatileState.PeerServers)
				Heartbeat(state.GetServerState().VolatileState.NewPeerServers)
			}
			state.HeartBeatChan <- "Heartbeat success"
		}
		time.Sleep(1000 * time.Millisecond)
		//logger.WithContext(ctx).Printf("number of goroutines:%d", runtime.NumGoroutine())
		state.PrintState(context.Background())
	}
}

func IamLeader(ctx context.Context) (bool, error) {
	if state.GetServerState().Role == raft.Role_Leader {
		if state.GetServerState().MemberConf.State == conf.COld {
			res, err := IamLeaderToServers(ctx, state.GetServerState().VolatileState.PeerServers)
			if err != nil {
				return false, err
			}
			return res, nil

		} else if state.GetServerState().MemberConf.State == conf.COldNew {
			res, err := IamLeaderToServers(ctx, state.GetServerState().VolatileState.PeerServers)
			if err != nil {
				return false, err
			}
			res2, err := IamLeaderToServers(ctx, state.GetServerState().VolatileState.NewPeerServers)
			if err != nil {
				return false, err
			}
			return res && res2, nil
		} else {
			return false, fmt.Errorf("memberconf state error")
		}
	} else {
		return false, nil
	}
}

func IamLeaderToServers(ctx context.Context, servers []string) (bool, error) {
	if len(servers) == 0 {
		return true, nil
	}
	var needCount = uint32(0)
	var res = true
	var once sync.Once
	var ch = make(chan int)

	for id, addr := range servers {
		addrTemp := addr
		idTemp := id
		req := &raft.AppendEntriesReq{
			Term:         state.GetServerState().PersistentState.CurrentTerm,
			LeaderId:     int64(state.GetServerState().ServerId),
			LeaderCommit: state.GetServerState().VolatileState.CommitIndex,
		}
		go func() {
			var err error
			defer func() {
				if err == nil {
					atomic.AddUint32(&needCount, 1)
					if needCount+1 >= state.GetMaxNum(len(servers)) {
						once.Do(func() {
							ch <- 1
						})
					}
				}
			}()
			logger.WithContext(ctx).Debugf("leader:%d send Heartbeat to follower:%d，req:%v\n", state.GetServerState().ServerId, idTemp, req)
			var resp *raft.AppendEntriesResp
			resp, err = rpc.AppendEntriesByServer(ctx, idTemp, addrTemp, req, false)
			if err != nil {
				logger.WithContext(ctx).Errorf("heart beat to %d err: %v", idTemp, err)
				return
			}
			if resp.Term > state.GetServerState().PersistentState.CurrentTerm {
				state.ToFollower(resp.Term)
				res = false
				once.Do(func() {
					ch <- 1
				})
			}
		}()
	}
	select {
	case <-time.After(1 * time.Second):
		return false, fmt.Errorf("timeout err")
	case <-ch:
		return res, nil
	}
}

func Heartbeat(server []string) {

	for id, addr := range server {
		addrTemp := addr
		idTemp := id
		req := &raft.AppendEntriesReq{
			Term:         state.GetServerState().PersistentState.CurrentTerm,
			LeaderId:     int64(state.GetServerState().ServerId),
			LeaderCommit: state.GetServerState().VolatileState.CommitIndex,
		}
		go func() {
			logger.Debugf("leader:%d send Heartbeat to follower:%d，req:%v\n", state.GetServerState().ServerId, idTemp, req)
			resp, err := rpc.AppendEntriesByServer(context.Background(), idTemp, addrTemp, req, false)
			if err != nil {
				logger.Errorf("heart beat to %d err: %v", idTemp, err)
				return
			}
			if resp.Term > state.GetServerState().PersistentState.CurrentTerm {
				state.ToFollower(resp.Term)
			}
		}()
	}
}
