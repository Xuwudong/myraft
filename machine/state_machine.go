package machine

import (
	"github.com/Xuwudong/myraft/state"
	"log"
	"sync"
)

var stateMachine = &StateMachine{
	KVMap: make(map[string]int64, 0),
}

var once sync.Once

type StateMachine struct {
	KVMap map[string]int64
}

func GetStateMachine() *StateMachine {
	return stateMachine
}

func Apply() error {
	err := Init()
	if err != nil {
		return err
	}
	ApplyFromIndex(int(state.GetServerState().VolatileState.CommitIndex) + 1)
	return nil
}

func Init() error {
	once.Do(func() {
		ApplyFromIndex(0)
		log.Println("init state machine finished")
	})
	return nil
}

func ApplyFromIndex(i int) {
	for ; i < len(state.GetServerState().PersistentState.Logs); i++ {
		entity := state.GetServerState().PersistentState.Logs[i].Command.Entity
		stateMachine.KVMap[entity.Key] = entity.Value
	}
	state.SetVolatileState(int64(i-1), int64(i-1))
}
