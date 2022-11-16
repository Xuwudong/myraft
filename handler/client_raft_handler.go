package handler

import (
	"_9932xt/myraft/gen-go/raft"
	log2 "_9932xt/myraft/log"
	"_9932xt/myraft/machine"
	"_9932xt/myraft/service"
	"_9932xt/myraft/state"
	"context"
	"log"
)

type ClientRaftHandler struct {
}

func NewClientRaftHandler() *ClientRaftHandler {
	return &ClientRaftHandler{}
}

func (p *ClientRaftHandler) DoCommand(ctx context.Context, req *raft.DoCommandReq) (*raft.DoCommandResp, error) {
	resp := &raft.DoCommandResp{}
	defer func() {
		log.Printf("DoCommand req:%v\n", req)
		log.Printf("DoCommand resp:%v\n", resp)
	}()
	if req.Command == nil {
		return resp, nil
	}
	if req.Command.Opt == raft.Opt_Write {
		if state.GetServerState().Role != raft.Role_Leader {
			return resp, nil
		}
		// todo ID去重
		logEntry := &raft.LogEntry{
			Term:    state.GetServerState().PersistentState.CurrentTerm,
			Command: req.Command,
		}
		resp = &raft.DoCommandResp{}
		_, _, err := log2.AppendLog([]*raft.LogEntry{logEntry})
		if err != nil {
			log.Println(err)
			return resp, nil
		}
		// send logEntry
		req := &raft.AppendEntriesReq{
			Term:         state.GetServerState().PersistentState.CurrentTerm,
			LeaderId:     int64(state.GetServerState().ServerId),
			LeaderCommit: state.GetServerState().VolatileState.CommitIndex,
		}
		err = service.AppendLogEntries(ctx, req)
		if err != nil {
			log.Printf("AppendLogEntries error:%v\n\n", err)
			return resp, nil
		}
		// reply to status machine
		err = machine.Apply()
		if err != nil {
			log.Printf("machine apply error:%v\n\n", err)
			return resp, err
		}
		resp.Succuess = true
	} else if req.Command.Opt == raft.Opt_Read {
		//todo 第二，领导人在处理只读的请求之前必须检查自己是否已经被废黜了（他自己的信息已经变脏了如果一个更新的领导人被选举出来）。
		//Raft 中通过让领导人在响应只读请求之前，先和集群中的大多数节点交换一次心跳信息来处理这个问题。
		if req.Command.Entity == nil {
			return resp, nil
		}
		if state.GetServerState().Role == raft.Role_Follower {
			err := machine.Apply()
			if err != nil {
				return resp, err
			}
		}
		value, _ := machine.GetStateMachine().KVMap[req.Command.Entity.Key]
		resp.Value = value
		resp.Succuess = true
	}
	return resp, nil
}
