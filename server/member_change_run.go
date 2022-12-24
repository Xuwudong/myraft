package server

import (
	"context"
	"github.com/Xuwudong/myraft/conf"
	"github.com/Xuwudong/myraft/gen-go/raft"
	log2 "github.com/Xuwudong/myraft/log"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/service"
	"github.com/Xuwudong/myraft/state"
	"time"
)

func ToNewServerAddrMapLoop() {
	for {
		select {
		case <-state.ToNewServerAddrMapChannel:
			for {
				err := AppendCNewConf()
				if err != nil {
					logger.WithContext(context.Background()).Errorf("appendCNewConf error:%v", err)
					continue
				}
				break
			}
		case <-time.After(1 * time.Hour):
		}
	}
}

func AppendCNewConf() error {
	if state.GetServerState().MemberConf.State != conf.COld {
		logEntry := &raft.LogEntry{
			Term: state.GetServerState().PersistentState.CurrentTerm,
			Entry: &raft.Entry{
				EntryType: raft.EntryType_MemberChangeNew,
			},
		}
		_, _, err := log2.AppendLog(context.Background(), []*raft.LogEntry{logEntry})
		if err != nil {
			logger.WithContext(context.Background()).Println(err)
			return err
		}
		state.ToCOldState(context.Background())
	}
	// send logEntry
	req := &raft.AppendEntriesReq{
		Term:         state.GetServerState().PersistentState.CurrentTerm,
		LeaderId:     int64(state.GetServerState().ServerId),
		LeaderCommit: state.GetServerState().VolatileState.CommitIndex,
	}
	err := service.AppendLogEntriesToMost(context.Background(), req)
	if err != nil {
		logger.WithContext(context.Background()).Errorf("AppendLogEntriesToMost error:%v\n\n", err)
		return err
	}
	return nil
}
