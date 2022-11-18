package server

import (
	"context"
	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/net"
	"github.com/Xuwudong/myraft/rpc"
	"github.com/Xuwudong/myraft/state"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
)

func Run() {
	rand.Seed(time.Now().Unix())
	ranTime := rand.Intn(5000) + 5000
	timeout, _ := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(ranTime))
	for {
		select {
		case res := <-state.HeartBeatChan:
			log.Printf("receive msg:%s\n", res)
			// 重置定时器
			ranTime = rand.Intn(5000) + 5000
			timeout, _ = context.WithTimeout(context.Background(), time.Millisecond*time.Duration(ranTime))
		case <-timeout.Done():
			// 开始新的超时周期
			ranTime = rand.Intn(5000) + 5000
			timeout, _ = context.WithTimeout(context.Background(), time.Millisecond*time.Duration(ranTime))

			// 超时了,发起选举
			state.GetServerState().Role = raft.Role_Candidater
			err := state.SetTerm(int(state.GetServerState().PersistentState.CurrentTerm + 1))
			if err != nil {
				log.Printf("setTerm error:%v", err)
				continue
			}
			log.Printf("start election,serverId: %d, term:%d", state.GetServerState().ServerId, state.GetServerState().PersistentState.CurrentTerm)

			lastLogTerm, lastLogIndex := state.GetLastLogMsg()

			requestVoteReq := &raft.RequestVoteReq{
				Term:         state.GetServerState().PersistentState.CurrentTerm,
				CandidateId:  int64(state.GetServerState().ServerId),
				LastLogIndex: lastLogIndex,
				LastLogTerm:  lastLogTerm,
			}
			// 初始化选票
			atomic.StoreUint32(&state.GetServerState().VolatileState.VoteNum, 0)
			// 自己先得一票
			atomic.AddUint32(&state.GetServerState().VolatileState.VoteNum, 1)
			var wg sync.WaitGroup
			for _, addr := range state.GetServerState().Conf.InnerAddrMap {
				if addr != state.GetServerState().Net.ServerAddr {
					wg.Add(1)
					addrTemp := addr
					go func() {
						defer wg.Done()
						var (
							client    *raft.RaftServerClient
							transport thrift.TTransport
						)
						retry := 0
						for {
							client, transport, err = rpc.NewClient(net.TransportFactory, net.ProtocolFactory, addrTemp, net.Secure, net.Cfg)
							if err == nil {
								break
							}
							log.Printf("error new client: %v", err)
							retry++
							if retry > 5 {
								break
							}
						}
						if client != nil {
							resp, err := rpc.RequestVote(client, context.Background(), requestVoteReq)
							if err != nil {
								log.Printf("requestVote error:%v", err)
								return
							}
							defer func(transport thrift.TTransport) {
								err := transport.Close()
								if err != nil {
									log.Printf("transport error: %v", err)
								}
							}(transport)
							if resp.Term > state.GetServerState().PersistentState.CurrentTerm {
								state.ToFollower(resp.Term)
								return
							}
							if resp.VoteGranted {
								atomic.AddUint32(&state.GetServerState().VolatileState.VoteNum, 1)
							}
						}
					}()
				}
			}
			wg.Wait()
			log.Printf("actual:%d, expect:%d", state.GetServerState().VolatileState.VoteNum, state.GetMaxNum())
			if state.GetServerState().VolatileState.VoteNum >= state.GetMaxNum() {
				state.GetServerState().Role = raft.Role_Leader
				log.Printf("i am coming to leader:%d\n", state.GetServerState().ServerId)
				// 广播master心跳
				appendEntriesReq := &raft.AppendEntriesReq{
					Term:     state.GetServerState().PersistentState.CurrentTerm,
					LeaderId: int64(state.GetServerState().ServerId),
				}
				var wg sync.WaitGroup
				for id, addr := range state.GetServerState().Conf.InnerAddrMap {
					if addr != state.GetServerState().Net.ServerAddr {
						wg.Add(1)
						addrTemp := addr
						tempId := id
						go func() {
							defer wg.Done()
							resp, err := rpc.AppendEntriesByServer(tempId, addrTemp, appendEntriesReq, false)
							if err != nil {
								log.Printf(" AppendEntries error:%v", err)
								return
							}
							if resp.Term > state.GetServerState().PersistentState.CurrentTerm {
								// 广播master心跳失败,可能是因为已经有新的master被选出，这里网络分区情况下可能导致算法不工作
								// 假如有a,b,c 三个实例，b是主，a收不到b的心跳会发起选举投票，c收到后将自己状态切换为candidator并发起投票
								// b收到c的投票后也切换成follower
								state.ToFollower(resp.Term)
							}
						}()
					}
				}
				wg.Wait()
				log.Printf("notify followers i am leader success")
				if state.GetServerState().Role == raft.Role_Leader {
					log.Printf("send empty command start")
					go func() {
						state.HeartBeatChan <- "master selected"
					}()
					// 初始化nextIndexMap,matchIndexMap
					state.InitMasterVolatileState()
					// 领导人完全特性保证了领导人一定拥有所有已经被提交的日志条目，但是在他任期开始的时候，他可能不知道哪些是
					//已经被提交的。为了知道这些信息，他需要在他的任期里提交一条日志条目。Raft 中通过领导人在任期开始的时
					//候提交一个空白的没有任何操作的日志条目到日志中去来实现。
					for {
						resp, err := clientRaftHandler.DoCommand(context.Background(), &raft.DoCommandReq{
							Command: &raft.Command{
								Opt:    raft.Opt_Write,
								Entity: &raft.Entity{},
							},
						})
						if err != nil {
							log.Printf("send empty command error:%v\n", err)
							continue
						}
						if !resp.Succuess {
							continue
						}
						log.Printf("send empty command success")
						//atomic.StoreUint32(&state.GetServerState().VolatileState.VoteNum, 0)
						break
					}
				}
			}
		}
	}
}
