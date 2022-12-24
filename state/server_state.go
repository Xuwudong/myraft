package state

import (
	"context"
	"encoding/json"
	"github.com/Xuwudong/myraft/conf"
	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/net"
	"github.com/Xuwudong/myraft/pool"
	"strconv"
	"sync"
)

var HeartBeatChan = make(chan string)

var serverState = &ServerState{
	Role:            raft.Role_Follower,
	Net:             &net.Net{},
	PersistentState: &PersistentState{},
	VolatileState: &VolatileState{
		CommitIndex: int64(-1),
		LastApplied: int64(-1),
	},
	MasterVolatileState: &MasterVolatileState{},
	SlaveVolatileState:  &SlaveVolatileState{},
	MemberConf: &conf.MemberConf{
		ServerAddrMap:    map[int64]string{},
		ClientAddrMap:    map[int64]string{},
		NewServerAddrMap: map[int64]string{},
	},
}

type ServerState struct {
	ServerId            int
	Role                raft.Role
	Net                 *net.Net
	PersistentState     *PersistentState
	VolatileState       *VolatileState
	MasterVolatileState *MasterVolatileState
	SlaveVolatileState  *SlaveVolatileState
	Conf                *conf.Conf
	MemberConf          *conf.MemberConf
}

type PersistentState struct {
	// 服务器已知最新的任期（在服务器首次启动时初始化为0，单调递增）
	CurrentTerm int64
	// 当前任期内收到选票的 candidateId，如果没有投给任何候选人 则为空
	VotedFor *VotedFor
	// 日志条目；每个条目包含了用于状态机的命令，以及领导人接收到该条目时的任期（初始索引为1）
	Logs []*raft.LogEntry
}

type VotedFor struct {
	Term      int64
	Candidate int
}

type VolatileState struct {
	// 已知已提交的最高的日志条目的索引（初始值为0，单调递增）
	CommitIndex int64
	// 已经被应用到状态机的最高的日志条目的索引（初始值为0，单调递增）
	LastApplied int64
	// 选自己的票数
	VoteNum    uint32
	NewVoteNum uint32
	// peerServers
	PeerServers    []string
	NewPeerServers []string
}

type MasterVolatileState struct {
	// 对于每一台服务器，发送到该服务器的下一个日志条目的索引（初始值为领导人最后的日志条目的索引+1）
	NextIndexMap sync.Map
	// 对于每一台服务器，已知的已经复制到该服务器的最高日志条目的索引（初始值为0，单调递增）
	MatchIndexMap sync.Map
}

type SlaveVolatileState struct {
	LeaderId int64
}

func GetServerState() *ServerState {
	return serverState
}

func GetLastLogMsg() (int64, int64) {
	lastLogIndex := 0
	lastLogTerm := int64(0)
	if len(GetServerState().PersistentState.Logs) > 0 {
		lastLogIndex = len(GetServerState().PersistentState.Logs) - 1
		lastLogTerm = GetServerState().PersistentState.Logs[len(GetServerState().PersistentState.Logs)-1].Term
	}
	return lastLogTerm, int64(lastLogIndex)
}

func ToFollower(term int64) {
	err := SetTerm(int(term))
	if err != nil {
		logger.Errorf("set Term err:%v", err)
	}
	GetServerState().Role = raft.Role_Follower
}

var masterVolatileStateLock sync.Mutex

func InitMasterVolatileState() {
	masterVolatileStateLock.Lock()
	defer masterVolatileStateLock.Unlock()
	if GetServerState().MemberConf.State == conf.COld {
		for id, _ := range GetServerState().VolatileState.PeerServers {
			GetServerState().MasterVolatileState.NextIndexMap.Store(id, int64(len(GetServerState().PersistentState.Logs)))
			GetServerState().MasterVolatileState.MatchIndexMap.Store(id, -1)
		}
	} else if GetServerState().MemberConf.State == conf.COldNew {
		for id, _ := range GetServerState().VolatileState.PeerServers {
			GetServerState().MasterVolatileState.NextIndexMap.Store(id, int64(len(GetServerState().PersistentState.Logs)))
			GetServerState().MasterVolatileState.MatchIndexMap.Store(id, -1)
		}
		for id, _ := range GetServerState().VolatileState.NewPeerServers {
			GetServerState().MasterVolatileState.NextIndexMap.Store(id, int64(len(GetServerState().PersistentState.Logs)))
			GetServerState().MasterVolatileState.MatchIndexMap.Store(id, -1)
		}
	}
}

func SetMasterVolatileState(id int, nextIndex, matchIndex int64) {
	masterVolatileStateLock.Lock()
	defer masterVolatileStateLock.Unlock()
	GetServerState().MasterVolatileState.NextIndexMap.Store(id, nextIndex)
	GetServerState().MasterVolatileState.MatchIndexMap.Store(id, matchIndex)
}

//var volatileStateLock sync.Mutex

//func SetVolatileState(commitIndex, lastApplied int64) {
//	volatileStateLock.Lock()
//	defer volatileStateLock.Unlock()
//	GetServerState().VolatileState.CommitIndex = commitIndex
//	GetServerState().VolatileState.LastApplied = lastApplied
//}

// GetMaxNum 大多数选票计算方式
func GetMaxNum(peerServers int) uint32 {
	return uint32((peerServers+1)/2 + 1)
}

func SetVotedFor(term int, candidate int) error {
	GetServerState().PersistentState.VotedFor = &VotedFor{
		Term:      int64(term),
		Candidate: candidate,
	}
	dateFile := GetServerState().Conf.LogDir + strconv.FormatInt(int64(GetServerState().ServerId), 10) + ".data"
	return conf.UpdateDataField(dateFile, conf.VotedFor, term, candidate)
}

func SetTerm(term int) error {
	GetServerState().PersistentState.CurrentTerm = int64(term)
	dateFile := GetServerState().Conf.LogDir + strconv.FormatInt(int64(GetServerState().ServerId), 10) + ".data"
	return conf.UpdateDataField(dateFile, conf.Term, term, 0)
}

var ToNewServerAddrMapChannel = make(chan int)

func ToNewServerAddrMap() {
	ToNewServerAddrMapChannel <- 1
}

func ToCOldNewState(ctx context.Context, entry *raft.Entry) {
	GetServerState().MemberConf.State = conf.COldNew
	addServerAdds := make([]string, 0)
	memberMap := make(map[int64]interface{})
	for _, member := range entry.Members {
		if _, ok := GetServerState().MemberConf.ServerAddrMap[member.MemberID]; !ok {
			addServerAdds = append(addServerAdds, member.ServerAddr)
		}
		memberMap[member.MemberID] = &struct{}{}
		GetServerState().MemberConf.NewServerAddrMap[member.MemberID] = member.ServerAddr
		GetServerState().MemberConf.ClientAddrMap[member.MemberID] = member.ClientAddr
	}
	pool.Init(addServerAdds)
	for id, _ := range GetServerState().MemberConf.ServerAddrMap {
		if _, ok := memberMap[id]; !ok {
			delete(GetServerState().MemberConf.ServerAddrMap, id)
			delete(GetServerState().MemberConf.ClientAddrMap, id)
		}
	}
	peerServers := make([]string, 0)
	for _, addr := range GetServerState().MemberConf.NewServerAddrMap {
		if addr != GetServerState().Net.ServerAddr {
			peerServers = append(peerServers, addr)
		}
	}
	GetServerState().VolatileState.NewPeerServers = peerServers
	PrintState(ctx)
}

func ToCOldState(ctx context.Context) {
	GetServerState().MemberConf.State = conf.COld

	GetServerState().MemberConf.ServerAddrMap = GetServerState().MemberConf.NewServerAddrMap
	GetServerState().MemberConf.NewServerAddrMap = make(map[int64]string)

	GetServerState().VolatileState.PeerServers = GetServerState().VolatileState.NewPeerServers
	GetServerState().VolatileState.NewPeerServers = nil
	PrintState(ctx)
}

func PrintState(ctx context.Context) {
	var liteState ServerState
	bytes, _ := json.Marshal(GetServerState())
	json.Unmarshal(bytes, &liteState)
	liteState.PersistentState.Logs = nil
	bytes, _ = json.Marshal(liteState)
	logger.WithContext(ctx).Printf("serverState:%+v", string(bytes))
}
