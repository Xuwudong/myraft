package handler

import (
	"context"
	"github.com/Xuwudong/myraft/heartbeat"
	"github.com/apache/thrift/lib/go/thrift"

	"github.com/Xuwudong/myraft/gen-go/raft"
	log2 "github.com/Xuwudong/myraft/log"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/machine"
	"github.com/Xuwudong/myraft/service"
	"github.com/Xuwudong/myraft/state"
)

type ClientRaftHandler struct {
}

func NewClientRaftHandler() *ClientRaftHandler {
	return &ClientRaftHandler{}
}

func (p *ClientRaftHandler) DoCommand(ctx context.Context, req *raft.DoCommandReq) (*raft.DoCommandResp, error) {
	resp := &raft.DoCommandResp{}
	defer func() {
		logger.WithContext(ctx).Infof("DoCommand req:%v, resp:%v", req, resp)
	}()
	if req.Command == nil {
		return resp, nil
	}
	// 只有leader能响应读写请求
	if state.GetServerState().Role != raft.Role_Leader {
		leader := state.GetServerState().Conf.OuterAddrMap[int(state.GetServerState().SlaveVolatileState.LeaderId)]
		resp.Leader = &leader
		return resp, nil
	}
	if req.Command.Opt == raft.Opt_Write {
		// todo ID去重
		logEntry := &raft.LogEntry{
			Term:    state.GetServerState().PersistentState.CurrentTerm,
			Command: req.Command,
		}
		resp = &raft.DoCommandResp{}
		_, _, err := log2.AppendLog(ctx, []*raft.LogEntry{logEntry})
		if err != nil {
			logger.WithContext(ctx).Println(err)
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
			logger.WithContext(ctx).Errorf("AppendLogEntries error:%v\n\n", err)
			return resp, nil
		}
		// reply to status machine
		//err = machine.Apply(ctx)
		//if err != nil {
		//	logger.WithContext(ctx).Errorf("machine apply error:%v\n\n", err)
		//	return resp, err
		//}
		resp.Succuess = true
	} else if req.Command.Opt == raft.Opt_Read {
		if req.Command.Entity == nil {
			return resp, nil
		}
		//第二，领导人在处理只读的请求之前必须检查自己是否已经被废黜了（他自己的信息已经变脏了如果一个更新的领导人被选举出来）。
		//Raft 中通过让领导人在响应只读请求之前，先和集群中的大多数节点交换一次心跳信息来处理这个问题。
		res, err := heartbeat.IamLeader()
		if err != nil {
			logger.WithContext(ctx).Errorf("check iamleader error:%v", err)
			return resp, nil
		}
		if !res {
			logger.WithContext(ctx).Errorf("oh no, i am not leader at all")
			resp.Leader = thrift.StringPtr(state.GetServerState().Conf.OuterAddrMap[int(state.GetServerState().SlaveVolatileState.LeaderId)])
			return resp, nil
		}
		err = machine.Apply(ctx)
		if err != nil {
			return resp, err
		}
		value, _ := machine.GetStateMachine().KVMap.Load(req.Command.Entity.Key)
		v, _ := value.(int64)
		if v == 0 {
			logger.WithContext(ctx).Errorf("read error no value, key:%s", req.Command.Entity.Key)
		}
		resp.Value = v
		resp.Succuess = true
	}
	return resp, nil
}
