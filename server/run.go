package server

import (
	"_9932xt/myraft/gen-go/raft"
	"_9932xt/myraft/rpc"
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func Run() {
	ranTime := rand.Intn(150) + 150
	ch := make(chan string)
	timeout, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(ranTime))
	defer cancel()
	for {
		select {
		case res := <-ch:
			fmt.Printf("receive msg:%s\n", res)
		case <-timeout.Done():
			// 开始新的超时周期
			timeout, cancel = context.WithTimeout(context.Background(), time.Millisecond*time.Duration(ranTime))

			// 超时了,发起选举
			GetServer().Role = raft.Role_Candidater
			GetServer().PersistentState.CurrentTerm += 1

			lastLogIndex := 0
			lastLogTerm := int64(0)
			if len(GetServer().PersistentState.Logs) > 0 {
				lastLogIndex = len(GetServer().PersistentState.Logs) - 1
				lastLogTerm = GetServer().PersistentState.Logs[len(GetServer().PersistentState.Logs)-1].Term
			}

			requestVoteReq := &raft.RequestVoteReq{
				Term:         GetServer().PersistentState.CurrentTerm,
				CandidateId:  int64(GetServer().ServerId),
				LastLogIndex: int64(lastLogIndex),
				LastLogTerm:  lastLogTerm,
			}
			// 自己先得一票
			atomic.AddUint32(&GetServer().VolatileState.VoteNum, 1)
			for _, ps := range GetServer().Net.PeerServerMap {
				curPs := ps
				go func() {
					resp, err := rpc.RequestVote(curPs.Client, context.Background(), requestVoteReq)
					if err != nil {
						log.Printf("requestVote error:%v", err)
						return
					}
					if resp.Term > GetServer().PersistentState.CurrentTerm {
						GetServer().PersistentState.CurrentTerm = resp.Term
						GetServer().Role = raft.Role_Follower
					}
					if resp.VoteGranted {
						atomic.AddUint32(&GetServer().VolatileState.VoteNum, 1)
					}
					if GetServer().VolatileState.VoteNum >= getMaxNum() {
						GetServer().Role = raft.Role_Leader
						// 广播master心跳
						appendEntriesReq := &raft.AppendEntriesReq{
							Term:     GetServer().PersistentState.CurrentTerm,
							LeaderId: int64(GetServer().ServerId),
						}
						var wg sync.WaitGroup
						for _, ps := range GetServer().Net.PeerServerMap {
							wg.Add(1)
							curPs := ps
							go func() {
								defer wg.Done()
								resp, err := rpc.AppendEntries(curPs.Client, context.Background(), appendEntriesReq)
								if err != nil {
									log.Printf(" AppendEntries error:%v", err)
									return
								}
								if resp.Term > GetServer().PersistentState.CurrentTerm {
									// 广播master心跳失败,可能是因为已经有新的master被选出，这里网络分区情况下可能导致算法不工作
									// 假如有a,b,c 三个实例，b是主，a收不到b的心跳会发起选举投票，c收到后将自己状态切换为candidator并发起投票
									// b收到c的投票后也切换成follower
									GetServer().PersistentState.CurrentTerm = resp.Term
									GetServer().Role = raft.Role_Follower
								}
							}()
						}
						wg.Wait()
						if GetServer().Role == raft.Role_Leader {
							//todo 领导人完全特性保证了领导人一定拥有所有已经被提交的日志条目，但是在他任期开始的时候，他可能不知道哪些是
							//已经被提交的。为了知道这些信息，他需要在他的任期里提交一条日志条目。Raft 中通过领导人在任期开始的时
							//候提交一个空白的没有任何操作的日志条目到日志中去来实现。
							ch <- "done"
						}
					}
				}()
			}
		}
	}

}

// 大多数选票计算方式
func getMaxNum() uint32 {
	total := len(GetServer().Net.PeerServerMap)
	return uint32(total/2 + 1)
}
