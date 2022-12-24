package pool

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	"github.com/apache/thrift/lib/go/thrift"
	"sync"
)

type Client struct {
	lock sync.Mutex
	*raft.RaftServerClient
	inUse bool
	thrift.TTransport
	server string
	close  bool
}

var clientPool = make(map[string][]*Client, 0)

func Init(servers []string) {
	for _, addr := range servers {
		clientPool[addr] = make([]*Client, 0)
		for i := 0; i < 200; i++ {
			client, tr, err := newClient(addr)
			if err != nil {
				logger.Errorf("new Client error:%v", err)
				continue
			}
			clientPool[addr] = append(clientPool[addr], &Client{
				RaftServerClient: client,
				TTransport:       tr,
				server:           addr,
			})
		}
		logger.Infof("init client poor success, addr:%s", addr)
	}
}

func getClientByServer(server string) (*Client, error) {
	clients, ok := clientPool[server]
	if !ok {
		return nil, fmt.Errorf("invalid server:%s", server)
	}
	for _, client := range clients {
		if useClient(client) {
			return client, nil
		}
	}
	client, tf, err := newClient(server)
	if err != nil {
		return nil, fmt.Errorf("new client err:%v", err)
	}
	if len(clients) > 1000 {
		return nil, fmt.Errorf("no valid client")
	}
	c := &Client{
		RaftServerClient: client,
		TTransport:       tf,
		server:           server,
	}
	clients = append(clients, c)
	if useClient(c) {
		return c, nil
	}
	return nil, fmt.Errorf("no valid client")
}

func (c *Client) recycle() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.inUse = false
	return nil
}

func (c *Client) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.close = true
	clients := clientPool[c.server]
	deleteClient(clients, c)
	clientPool[c.server] = clients
	return c.TTransport.Close()
}

func AppendEntries(ctx context.Context, server string, req *raft.AppendEntriesReq) (*raft.AppendEntriesResp, error) {
	var (
		client *Client
		err    error
	)
	client, err = getClientByServer(server)
	if err != nil {
		return nil, err
	}
	logger.Errorf("error get client: %v", err)
	defer func(client *Client) {
		err := client.recycle()
		if err != nil {
			logger.Errorf("Recycle error:%v", err)
		}
	}(client)
	res, err := client.RaftServerClient.AppendEntries(ctx, req)
	if err != nil {
		logger.WithContext(ctx).Errorf("Error during AppendEntries:%v", err)
		return nil, err
	}
	return res, err
}

func RequestVote(ctx context.Context, server string, req *raft.RequestVoteReq) (*raft.RequestVoteResp, error) {
	var (
		client *Client
		err    error
	)
	client, err = getClientByServer(server)
	if err != nil {
		return nil, err
	}
	logger.Errorf("error get client: %v", err)
	defer func(client *Client) {
		err := client.recycle()
		if err != nil {
			logger.Errorf("Recycle error:%v", err)
		}
	}(client)
	res, err := client.RaftServerClient.RequestVote(ctx, req)
	if err != nil {
		logger.WithContext(ctx).Errorf("Error during RequestVote:%v", err)
	}
	return res, err
}

func deleteClient(arr []*Client, c *Client) []*Client {
	j := 0
	for _, v := range arr {
		if v != c {
			arr[j] = v
			j++
		}
	}
	return arr[:j]
}

func useClient(client *Client) bool {
	client.lock.Lock()
	defer client.lock.Unlock()
	if !client.inUse && !client.close {
		client.inUse = true
		return true
	}
	return false
}

func newClient(server string) (*raft.RaftServerClient, thrift.TTransport, error) {
	var client *raft.RaftServerClient
	var err error
	var tr thrift.TTransport
	var count = 0
	for {
		client, tr, err = newServerClient(thrift.NewTTransportFactory(), thrift.NewTBinaryProtocolFactoryConf(nil), server,
			false, &thrift.TConfiguration{
				TLSConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			})
		if err == nil {
			break
		}
		count++
		logger.Printf("error new Client: %v", err)
		if count > 1 {
			return nil, tr, err
		}
	}
	return client, tr, err
}

func newServerClient(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string,
	secure bool, cfg *thrift.TConfiguration) (*raft.RaftServerClient, thrift.TTransport, error) {
	var transport thrift.TTransport
	if secure {
		transport = thrift.NewTSSLSocketConf(addr, cfg)
	} else {
		transport = thrift.NewTSocketConf(addr, cfg)
	}
	transport, err := transportFactory.GetTransport(transport)
	if err != nil {
		return nil, nil, err
	}
	//defer transport.Close()
	if err := transport.Open(); err != nil {
		return nil, nil, err
	}
	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)
	return raft.NewRaftServerClient(thrift.NewTStandardClient(iprot, oprot)), transport, nil
}
