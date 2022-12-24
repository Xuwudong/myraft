package pool

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReturn(t *testing.T) {
	server := []string{"localhost:9092"}
	Init(server)
	for i := 0; i < 200; i++ {
		ti := i
		go func() {
			fmt.Println(ti)
			c, err := getClientByServer(server[0])
			assert.Nil(t, err)
			err = c.recycle()
			assert.Nil(t, err)
		}()
		//time.Sleep(2 * time.Millisecond)
	}
	time.Sleep(10 * time.Second)
}
