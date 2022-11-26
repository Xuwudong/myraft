package heartbeat

import (
	"context"
	_ "net/http/pprof"
	"time"

	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/rpc"
	"github.com/Xuwudong/myraft/state"
)

func Run() {
	for {
		if state.GetServerState().Role == raft.Role_Leader {
			for id, addr := range state.GetServerState().Conf.InnerAddrMap {
				if addr != state.GetServerState().Net.ServerAddr {
					addrTemp := addr
					idTemp := id
					req := &raft.AppendEntriesReq{
						Term:         state.GetServerState().PersistentState.CurrentTerm,
						LeaderId:     int64(state.GetServerState().ServerId),
						LeaderCommit: state.GetServerState().VolatileState.CommitIndex,
						//PreLogIndex:  int64(len(state.GetServerState().PersistentState.Logs) - 1),
						//PreLogTerm:   state.GetServerState().PersistentState.Logs[int64(len(state.GetServerState().PersistentState.Logs)-1)].Term,
					}
					go func() {
						logger.Debugf("leader:%d send heartbeat to follower:%d，req:%v\n", state.GetServerState().ServerId, idTemp, req)
						resp, err := rpc.AppendEntriesByServer(context.Background(), idTemp, addrTemp, req, false)
						if err != nil {
							//e, ok := err.(*errno.NewClientErr)
							//if ok && e != nil {
							//	//delete(state.GetServerState().Net.PeerServerMap, idTemp)
							//}
							logger.Errorf("heart beat to %d err: %v", idTemp, err)
							return
						}
						if resp.Term > state.GetServerState().PersistentState.CurrentTerm {
							state.ToFollower(resp.Term)
						}
						state.HeartBeatChan <- "heartbeat success"
					}()
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
		//logger.WithContext(ctx).Printf("number of goroutines:%d", runtime.NumGoroutine())
		//bytes, _ := json.Marshal(state.GetServerState())
		//logger.Printf("serverState:%s", string(bytes))
	}
}
