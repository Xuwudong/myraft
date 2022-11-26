package pool

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRecycle(t *testing.T) {
	server := []string{"localhost:9092"}
	Init(server)
	for i := 0; i < 5000; i++ {
		ti := i
		go func() {
			fmt.Println(ti)
			c, err := GetClientByServer(server[0])
			assert.Nil(t, err)
			err = Recycle(c)
			assert.Nil(t, err)
		}()
		//time.Sleep(2 * time.Millisecond)
	}
	time.Sleep(1 * time.Second)
}
