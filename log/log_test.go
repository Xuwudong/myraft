package log

import (
	"_9932xt/myraft/gen-go/raft"
	"fmt"
	"testing"
)

func TestLogParse(t *testing.T) {
	log := &Log{
		Command: &raft.Command{
			Entry: &raft.Entry{
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
