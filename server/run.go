package server

import (
	"context"
	"github.com/Xuwudong/myraft/conf"
	"github.com/Xuwudong/myraft/heartbeat"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/rpc"
	"github.com/Xuwudong/myraft/state"
)

func Run() {
	rand.Seed(time.Now().Unix())
	ranTime := rand.Intn(3000) + 2000
	timeout, _ := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(ranTime))
	for {
		select {
		case res := <-state.HeartBeatChan:
			logger.Debugf("receive msg:%s\n", res)
			// 重置定时器
			ranTime = rand.Intn(3000) + 5000
			timeout, _ = context.WithTimeout(context.Background(), time.Millisecond*time.Duration(ranTime))
		case <-timeout.Done():
			// 开始新的超时周期
			ranTime = rand.Intn(3000) + 5000
			timeout, _ = context.WithTimeout(context.Background(), time.Millisecond*time.Duration(ranTime))

			// 超时了,发起选举
			state.GetServerState().Role = raft.Role_Candidater
			err := state.SetTerm(int(state.GetServerState().PersistentState.CurrentTerm + 1))
			if err != nil {
				logger.Errorf("setTerm error:%v", err)
				continue
			}
			logger.Infof("start election,serverId: %d, term:%d", state.GetServerState().ServerId, state.GetServerState().PersistentState.CurrentTerm)

			lastLogTerm, lastLogIndex := state.GetLastLogMsg()

			requestVoteReq := &raft.RequestVoteReq{
				Term:         state.GetServerState().PersistentState.CurrentTerm,
				CandidateId:  int64(state.GetServerState().ServerId),
				LastLogIndex: lastLogIndex,
				LastLogTerm:  lastLogTerm,
			}
			selectLeader(requestVoteReq)
		}
	}
}

func selectLeader(req *raft.RequestVoteReq) {
	flag := false
	if state.GetServerState().MemberConf.State == conf.COld {
		selectLeaderByServers(req, state.GetServerState().VolatileState.PeerServers, &state.GetServerState().VolatileState.VoteNum)
		if state.GetServerState().VolatileState.VoteNum >= state.GetMaxNum(len(state.GetServerState().VolatileState.PeerServers)) {
			flag = true
		}
	} else if state.GetServerState().MemberConf.State == conf.COldNew {
		selectLeaderByServers(req, state.GetServerState().VolatileState.PeerServers, &state.GetServerState().VolatileState.VoteNum)
		selectLeaderByServers(req, state.GetServerState().VolatileState.NewPeerServers, &state.GetServerState().VolatileState.NewVoteNum)
		if state.GetServerState().VolatileState.VoteNum >= state.GetMaxNum(len(state.GetServerState().VolatileState.PeerServers)) &&
			state.GetServerState().VolatileState.NewVoteNum >= state.GetMaxNum(len(state.GetServerState().VolatileState.NewPeerServers)) {
			flag = true
		}
	}
	if flag {
		state.GetServerState().Role = raft.Role_Leader
		logger.Infof("i am coming to leader:%d\n", state.GetServerState().ServerId)
		// 广播master心跳
		if state.GetServerState().MemberConf.State == conf.COld {
			heartbeat.Heartbeat(state.GetServerState().VolatileState.PeerServers)
		} else if state.GetServerState().MemberConf.State == conf.COldNew {
			heartbeat.Heartbeat(state.GetServerState().VolatileState.PeerServers)
			heartbeat.Heartbeat(state.GetServerState().VolatileState.NewPeerServers)
		}
		logger.Info("notify followers i am leader success")
		if state.GetServerState().Role == raft.Role_Leader {
			logger.Info("send empty command start")
			go func() {
				state.HeartBeatChan <- "master selected"
			}()
			// 初始化nextIndexMap,matchIndexMap
			state.InitMasterVolatileState()
			sendEmptyLogEntry()
		}
	}
}

func sendEmptyLogEntry() {
	// 领导人完全特性保证了领导人一定拥有所有已经被提交的日志条目，但是在他任期开始的时候，他可能不知道哪些是
	//已经被提交的。为了知道这些信息，他需要在他的任期里提交一条日志条目。Raft 中通过领导人在任期开始的时
	//候提交一个空白的没有任何操作的日志条目到日志中去来实现。
	for {
		resp, err := clientRaftHandler.DoCommand(context.Background(), &raft.DoCommandReq{
			Command: &raft.Command{
				Opt: raft.Opt_Write,
				Entry: &raft.Entry{
					EntryType: raft.EntryType_KV,
				},
			},
		})
		if err != nil {
			logger.Errorf("send empty command error:%v\n", err)
			continue
		}
		if !resp.Succuess {
			continue
		}
		logger.Printf("send empty command success")
		break
	}
}

func selectLeaderByServers(req *raft.RequestVoteReq, servers []string, voteNum *uint32) {
	//if len(servers) == 0 {
	//	//一台机器不给选举
	//	logger.Info("can not select leader when only has one server ")
	//	return
	//}
	atomic.StoreUint32(voteNum, 0)
	// 自己先得一票
	atomic.AddUint32(voteNum, 1)
	var wg sync.WaitGroup
	// 初始化选票
	for _, addr := range servers {
		if addr != state.GetServerState().Net.ServerAddr {
			wg.Add(1)
			addrTemp := addr
			go func() {
				defer wg.Done()
				resp, err := rpc.RequestVote(context.Background(), addrTemp, req)
				if err != nil {
					logger.Errorf("requestVote error:%v", err)
					return
				}
				if resp.Term > state.GetServerState().PersistentState.CurrentTerm {
					state.ToFollower(resp.Term)
					return
				}
				if resp.VoteGranted {
					atomic.AddUint32(voteNum, 1)
				}
			}()
		}
	}
	wg.Wait()
	logger.Infof("actual:%d, expect:%d", *voteNum, state.GetMaxNum(len(servers)))
}
