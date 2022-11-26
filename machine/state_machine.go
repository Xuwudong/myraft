package machine

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/state"
)

var stateMachine = &StateMachine{}

var once sync.Once

type StateMachine struct {
	KVMap sync.Map
}

func GetStateMachine() *StateMachine {
	return stateMachine
}

func Apply(ctx context.Context) error {
	err := Init(ctx)
	if err != nil {
		return err
	}
	if state.GetServerState().VolatileState.CommitIndex > state.GetServerState().VolatileState.LastApplied {
		ApplyFromIndex(ctx, state.GetServerState().VolatileState.LastApplied+1)
	}
	return nil
}

func Init(ctx context.Context) error {
	once.Do(func() {
		ApplyFromIndex(ctx, 0)
		logger.Info("init state machine finished")
	})
	return nil
}

func ApplyFromIndex(ctx context.Context, i int64) {
	var entity *raft.Entity
	for ; i <= state.GetServerState().VolatileState.CommitIndex && i < int64(len(state.GetServerState().PersistentState.Logs)); i++ {
		entity = state.GetServerState().PersistentState.Logs[i].Command.Entity
		stateMachine.KVMap.Store(entity.Key, entity.Value)
		logger.WithContext(ctx).Infof("apply index: %d, entity:%v", i, entity)
	}
	i--
	//logger.WithContext(ctx).Infof("apply to index: %d, entity:%v", i, entity)
	atomic.StoreInt64(&state.GetServerState().VolatileState.LastApplied, i)
}
