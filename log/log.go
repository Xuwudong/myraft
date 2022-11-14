package log

import (
	"_9932xt/myraft/gen-go/raft"
	"encoding/json"
	"strings"
)

type Log struct {
	Term    int64
	Command *raft.Command
}

func ParseLog(logLine string) (*Log, error) {
	log := &Log{}
	logLine = strings.Replace(logLine, "\n", "", -1)
	err := json.Unmarshal([]byte(logLine), log)
	if err != nil {
		return nil, err
	}
	return log, nil
}

func ToLogString(log *Log) (string, error) {
	bytes, err := json.Marshal(log)
	if err != nil {
		return string(bytes), err
	}
	return string(bytes) + "\n", nil
}
