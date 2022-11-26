package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/pool"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

func TestCommand(t *testing.T) {
	rand.Seed(time.Now().Unix())
	ran := rand.Int63n(10000)
	wReq := &raft.DoCommandReq{
		Command: &raft.Command{
			Opt: raft.Opt_Write,
			Entity: &raft.Entity{
				Key:   strconv.Itoa(int(ran)),
				Value: ran,
			},
		},
	}
	var client *raft.ClientRaftServerClient
	var err error
	var tr thrift.TTransport
	for {
		client, tr, err = pool.NewClientServerClient(thrift.NewTTransportFactory(), thrift.NewTBinaryProtocolFactoryConf(nil),
			"localhost:9090", false, &thrift.TConfiguration{
				TLSConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			})
		if err == nil {
			break
		}
		logger.Printf("error new client: %v", err)
	}
	defer tr.Close()
	resp, err := client.DoCommand(context.Background(), wReq)
	assert.Nil(t, err)
	assert.True(t, resp.Succuess == true)
	fmt.Println(resp)

	rReq := &raft.DoCommandReq{
		Command: &raft.Command{
			Opt: raft.Opt_Read,
			Entity: &raft.Entity{
				Key: strconv.Itoa(int(ran)),
			},
		},
	}
	resp, err = client.DoCommand(context.Background(), rReq)
	assert.Nil(t, err)
	assert.True(t, resp.Succuess == true)
	assert.True(t, resp.Value == ran)
	fmt.Println(resp)
}

func TestRead(t *testing.T) {
	var client *raft.ClientRaftServerClient
	var err error
	var tr thrift.TTransport
	for {
		client, tr, err = pool.NewClientServerClient(thrift.NewTTransportFactory(), thrift.NewTBinaryProtocolFactoryConf(nil),
			"localhost:9092", false, &thrift.TConfiguration{
				TLSConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			})
		if err == nil {
			break
		}
		logger.Printf("error new client: %v", err)
	}
	defer tr.Close()

	rReq := &raft.DoCommandReq{
		Command: &raft.Command{
			Opt: raft.Opt_Read,
			Entity: &raft.Entity{
				Key: "9765",
			},
		},
	}
	resp, err := client.DoCommand(context.Background(), rReq)
	assert.Nil(t, err)
	assert.True(t, resp.Succuess == true)
	assert.True(t, resp.Value == 9765)
	fmt.Println(resp)
}
