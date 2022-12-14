package handler

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"context"
	"github.com/Xuwudong/myraft/conf"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/Xuwudong/myraft/gen-go/raft"
	log2 "github.com/Xuwudong/myraft/log"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/state"
)

type RaftHandler struct {
}

func NewRaftHandler() *RaftHandler {
	return &RaftHandler{}
}

var appendMutex sync.Mutex

func (p *RaftHandler) AppendEntries(ctx context.Context, req *raft.AppendEntriesReq) (resp *raft.AppendEntriesResp, _err error) {
	defer func() {
		if len(req.Entries) > 0 {
			logger.WithContext(ctx).Printf("appendEntries req:%v, resp:%v", req, resp)
		}
	}()
	resp = &raft.AppendEntriesResp{
		Term: state.GetServerState().PersistentState.CurrentTerm,
	}
	if req.Term < state.GetServerState().PersistentState.CurrentTerm {
		resp.Succuess = false
	} else if req.Term > state.GetServerState().PersistentState.CurrentTerm {
		err := state.SetTerm(int(req.Term))
		if err != nil {
			return resp, err
		}
		resp.Term = state.GetServerState().PersistentState.CurrentTerm
	}
	state.GetServerState().Role = raft.Role_Follower
	if len(req.Entries) == 0 {
		// 心跳记得 ch <- "done",保证不超时
		state.HeartBeatChan <- "receive heartbeat"
		if state.GetServerState().SlaveVolatileState.LeaderId != req.LeaderId {
			state.GetServerState().SlaveVolatileState.LeaderId = req.LeaderId
		}
		//if state.GetServerState().VolatileState.CommitIndex < req.LeaderCommit {
		//	commitIndex := req.LeaderCommit
		//	if commitIndex >= int64(len(state.GetServerState().PersistentState.Logs)) {
		//		commitIndex = int64(len(state.GetServerState().PersistentState.Logs) - 1)
		//	}
		//	atomic.StoreInt64(&state.GetServerState().VolatileState.CommitIndex, commitIndex)
		//}
		resp.Succuess = true
	} else {
		appendMutex.Lock()
		defer appendMutex.Unlock()
		var match bool
		if req.PreLogIndex < 0 {
			match = true
		} else {
			if req.PreLogIndex < int64(len(state.GetServerState().PersistentState.Logs)) &&
				state.GetServerState().PersistentState.Logs[req.PreLogIndex].Term == req.PreLogTerm {
				match = true
			}
		}
		if !match {
			resp.Succuess = false
			return resp, nil
		}

		i := req.PreLogIndex + 1
		j := i
		deleted := false
		for _, entry := range req.Entries {
			if i > 0 && i < int64(len(state.GetServerState().PersistentState.Logs)) && state.GetServerState().PersistentState.Logs[i].Term != entry.Term {
				//如果一个已经存在的条目和新条目（译者注：即刚刚接收到的日志条目）发生了冲突（因为索引相同，任期不同），那么就删除这个已经存在的条目以及它之后的所有条目
				deleted = true
				logDir := state.GetServerState().Conf.LogDir
				logFilePath := logDir + strconv.FormatInt(int64(state.GetServerState().ServerId), 10) + ".log"

				err := log2.DeleteFrom(ctx, i, logFilePath)
				if err != nil {
					logger.WithContext(ctx).Errorf("deleteLogFrom %v", err)
					resp.Succuess = false
					return resp, nil
				}
				break
			}
			i++
		}
		newEntries := req.Entries
		if !deleted {
			newEntries = make([]*raft.LogEntry, 0)
			for _, entry := range req.Entries {
				if j >= 0 && j < int64(len(state.GetServerState().PersistentState.Logs)) {
					myLog := state.GetServerState().PersistentState.Logs[j]
					if myLog.Term == entry.Term &&
						myLog.Entry != nil && entry.Entry != nil &&
						myLog.Entry.Key == entry.Entry.Key &&
						myLog.Entry.Value == myLog.Entry.Value &&
						myLog.Entry.EntryType == myLog.Entry.EntryType {
						// 去重
						logger.WithContext(ctx).Infof("reduplicate log entry:%v", entry)
						j++
						continue
					}
				} else {
					newEntries = append(newEntries, entry)
					j++
				}
			}
		}
		for _, logEntry := range newEntries {
			if logEntry.Entry.EntryType == raft.EntryType_MemberChange {
				state.ToCOldNewState(ctx, logEntry.Entry)
			} else if logEntry.Entry.EntryType == raft.EntryType_MemberChangeNew {
				state.ToCOldState(ctx)
			}
		}
		_, _, err := log2.AppendLog(ctx, newEntries)
		if err != nil {
			logger.WithContext(ctx).Errorf("AppendLog %v", err)
			resp.Succuess = false
			return resp, nil
		}
		min := req.LeaderCommit
		if req.PreLogIndex+int64(len(req.Entries)) < req.LeaderCommit {
			min = req.PreLogIndex + int64(len(req.Entries))
		}
		atomic.StoreInt64(&state.GetServerState().VolatileState.CommitIndex, min)
		resp.Succuess = true
	}
	return resp, nil
}

var voteMutex sync.Mutex

func (p *RaftHandler) RequestVote(ctx context.Context, req *raft.RequestVoteReq) (*raft.RequestVoteResp, error) {
	resp := &raft.RequestVoteResp{
		Term: state.GetServerState().PersistentState.CurrentTerm,
	}
	defer func() {
		logger.WithContext(ctx).Printf("requestVote req:%v, resp:%v", req, resp)
	}()
	defer voteMutex.Unlock()
	voteMutex.Lock()
	if req.Term < state.GetServerState().PersistentState.CurrentTerm {
		resp.VoteGranted = false
		return resp, nil
	} else if req.Term > state.GetServerState().PersistentState.CurrentTerm {
		err := state.SetTerm(int(req.Term))
		if err != nil {
			return resp, err
		}
		resp.Term = state.GetServerState().PersistentState.CurrentTerm
		state.GetServerState().Role = raft.Role_Follower
	}
	if state.GetServerState().PersistentState.VotedFor == nil ||
		state.GetServerState().PersistentState.VotedFor.Term != req.Term ||
		state.GetServerState().PersistentState.VotedFor.Candidate == int(req.CandidateId) {
		lastLogTerm, lastLogIndex := state.GetLastLogMsg()
		if req.LastLogTerm > lastLogTerm || (req.LastLogTerm == lastLogTerm && req.LastLogIndex >= lastLogIndex) {
			resp.VoteGranted = true
			// 一旦同意了别人，对自己的选票置零
			if state.GetServerState().MemberConf.State == conf.COld {
				atomic.StoreUint32(&state.GetServerState().VolatileState.VoteNum, 0)
			} else if state.GetServerState().MemberConf.State == conf.COldNew {
				atomic.StoreUint32(&state.GetServerState().VolatileState.VoteNum, 0)
				atomic.StoreUint32(&state.GetServerState().VolatileState.NewVoteNum, 0)
			}
			err := state.SetVotedFor(int(req.Term), int(req.CandidateId))
			if err != nil {
				logger.WithContext(ctx).Errorf("SetVotedFor error:%v", err)
				resp.VoteGranted = false
			}
		}
	}
	return resp, nil
}
