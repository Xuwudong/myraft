package state

import (
	"_9932xt/myraft/conf"
	"_9932xt/myraft/gen-go/raft"
	"_9932xt/myraft/net"
	"log"
	"strconv"
	"sync"
)

var HeartBeatChan = make(chan string)

var serverState = &ServerState{
	Role:            raft.Role_Follower,
	Net:             &net.Net{},
	PersistentState: &PersistentState{},
	VolatileState:   &VolatileState{},
	MasterVolatileState: &MasterVolatileState{
		NextIndexMap:  make(map[int]int64, 0),
		MatchIndexMap: make(map[int]int64, 0),
	},
}

type ServerState struct {
	ServerId            int
	Role                raft.Role
	Net                 *net.Net
	PersistentState     *PersistentState
	VolatileState       *VolatileState
	MasterVolatileState *MasterVolatileState
	Conf                *conf.Conf
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
	VoteNum uint32
}

type MasterVolatileState struct {
	// 对于每一台服务器，发送到该服务器的下一个日志条目的索引（初始值为领导人最后的日志条目的索引+1）
	NextIndexMap map[int]int64
	// 对于每一台服务器，已知的已经复制到该服务器的最高日志条目的索引（初始值为0，单调递增）
	MatchIndexMap map[int]int64
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
		log.Printf("set Term err:%v", err)
	}
	GetServerState().Role = raft.Role_Follower
}

var masterVolatileStateLock sync.Mutex

func InitMasterVolatileState() {
	masterVolatileStateLock.Lock()
	defer masterVolatileStateLock.Unlock()
	for id, _ := range GetServerState().Conf.InnerAddrMap {
		if id != GetServerState().ServerId {
			GetServerState().MasterVolatileState.NextIndexMap[id] = int64(len(GetServerState().PersistentState.Logs))
			GetServerState().MasterVolatileState.MatchIndexMap[id] = 0
		}
	}
}

func SetMasterVolatileState(id int, nextIndex, matchIndex int64) {
	masterVolatileStateLock.Lock()
	defer masterVolatileStateLock.Unlock()
	GetServerState().MasterVolatileState.NextIndexMap[id] = nextIndex
	GetServerState().MasterVolatileState.MatchIndexMap[id] = matchIndex
}

var volatileStateLock sync.Mutex

func SetVolatileState(commitIndex, lastApplied int64) {
	volatileStateLock.Lock()
	defer volatileStateLock.Unlock()
	GetServerState().VolatileState.CommitIndex = commitIndex
	GetServerState().VolatileState.LastApplied = lastApplied
}

// GetMaxNum 大多数选票计算方式
func GetMaxNum() uint32 {
	total := len(GetServerState().Conf.InnerAddrMap)
	return uint32(total/2 + 1)
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
