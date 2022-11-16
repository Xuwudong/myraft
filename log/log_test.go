package log

import (
	"_9932xt/myraft/gen-go/raft"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogParse(t *testing.T) {
	log := &raft.LogEntry{
		Command: &raft.Command{
			Entity: &raft.Entity{
				Key:   "1",
				Value: 1,
			},
			Opt: raft.Opt_Write,
		},
		Term: 1,
	}
	str, err := ToLogString(log)
	fmt.Println(err)
	fmt.Println(str)
	log, err = ParseLog(str)
	fmt.Println(err)
	fmt.Println(log)

}

func TestDelete(t *testing.T) {
	logDir := "/tmp/myraft/"
	logFilePath := logDir + "1.log"

	err := DeleteFrom(11, logFilePath)
	assert.Nil(t, err)
}
