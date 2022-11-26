package log

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/state"
)

func ParseLog(logLine string) (*raft.LogEntry, error) {
	log := &raft.LogEntry{}
	logLine = strings.Replace(logLine, "\n", "", -1)
	err := json.Unmarshal([]byte(logLine), log)
	if err != nil {
		return nil, err
	}
	return log, nil
}

func ToLogString(log *raft.LogEntry) (string, error) {
	bytes, err := json.Marshal(log)
	if err != nil {
		return string(bytes), err
	}
	return string(bytes) + "\n", nil
}

func AppendLog(ctx context.Context, logs []*raft.LogEntry) (int64, int64, error) {
	var logsStr string
	for _, log := range logs {
		logStr, err := ToLogString(log)
		if err != nil {
			return 0, 0, err
		}
		logsStr += logStr
	}
	logDir := state.GetServerState().Conf.LogDir
	logFilePath := logDir + strconv.FormatInt(int64(state.GetServerState().ServerId), 10) + ".log"
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return 0, 0, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logger.WithContext(ctx).Error(err)
		}
	}(f)

	_, err = f.WriteString(logsStr)
	if err != nil {
		return 0, 0, err
	}
	preLogIndex := len(state.GetServerState().PersistentState.Logs) - 1
	var preLogTerm int64
	if preLogIndex >= 0 {
		preLogTerm = state.GetServerState().PersistentState.Logs[preLogIndex].Term
	}
	state.GetServerState().PersistentState.Logs = append(state.GetServerState().PersistentState.Logs, logs...)
	return int64(preLogIndex), preLogTerm, nil
}

func DeleteFrom(ctx context.Context, index int64, logFilePath string) error {
	lineBytes, err := ioutil.ReadFile(logFilePath)
	var lines []string
	if err != nil {
		logger.WithContext(ctx).Error(err)
		return err
	} else {
		contents := string(lineBytes)
		lines = strings.Split(contents, "\n")
	}
	var newLines []string
	for i, line := range lines {
		if line == "" {
			continue
		}
		if i < int(index) {
			newLines = append(newLines, line)
		}
	}

	state.GetServerState().PersistentState.Logs = state.GetServerState().PersistentState.Logs[0:index]

	f, err := os.OpenFile(logFilePath+".temp", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logger.WithContext(ctx).Error(err)
		}
	}(f)
	if len(newLines) > 0 {
		_, err = f.WriteString(strings.Join(newLines, "\n") + "\n")
	} else {
		_, err = f.WriteString(strings.Join(newLines, "\n"))
	}
	if err != nil {
		return err
	}
	err = os.Remove(logFilePath)
	if err != nil {
		return err
	}
	return os.Rename(logFilePath+".temp", logFilePath)
}
