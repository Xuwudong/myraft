package heartbeat

import (
	"_9932xt/myraft/gen-go/raft"
	"_9932xt/myraft/rpc"
	"_9932xt/myraft/state"
	"encoding/json"
	"log"
	_ "net/http/pprof"
	"runtime"
	"time"
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
					}
					go func() {
						log.Printf("leader:%d send heartbeat to follower:%d，req:%v\n", state.GetServerState().ServerId, idTemp, req)
						resp, err := rpc.AppendEntriesByServer(idTemp, addrTemp, req, false)
						if err != nil {
							//e, ok := err.(*errno.NewClientErr)
							//if ok && e != nil {
							//	//delete(state.GetServerState().Net.PeerServerMap, idTemp)
							//}
							log.Printf("heart beat to %d err: %v", idTemp, err)
							return
						}
						if resp.Term > state.GetServerState().PersistentState.CurrentTerm {
							state.ToFollower(resp.Term)
						}
					}()
				}
				state.HeartBeatChan <- "heartbeat success"
			}
			time.Sleep(2000 * time.Millisecond)
			log.Printf("number of goroutines:%d", runtime.NumGoroutine())
			bytes, _ := json.Marshal(state.GetServerState())
			log.Printf("serverState:%s\n", string(bytes))
		}
	}
}
