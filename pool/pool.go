package pool

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	thriftPool "github.com/Xuwudong/thrift-client-pool"
	"github.com/apache/thrift/lib/go/thrift"
)

func AppendEntries(ctx context.Context, server string, req *raft.AppendEntriesReq) (*raft.AppendEntriesResp, error) {
	tPool := MapPool.Get(server)
	ic, err := tPool.Get()
	if err != nil {
		logger.Errorf("error get client: %v", err)
		return nil, err
	}
	defer func(ic *thriftPool.IdleClient) {
		err := tPool.Put(ic)
		if err != nil {
			return
		}
	}(ic)
	res, err := ic.Client.(*raft.RaftServerClient).AppendEntries(ctx, req)
	if err != nil {
		logger.WithContext(ctx).Errorf("Error during AppendEntries:%v", err)
		return nil, err
	}
	return res, err
}

func RequestVote(ctx context.Context, server string, req *raft.RequestVoteReq) (*raft.RequestVoteResp, error) {
	tPool := MapPool.Get(server)
	ic, err := tPool.Get()
	if err != nil {
		logger.Errorf("error get client: %v", err)
		return nil, err
	}
	defer func(ic *thriftPool.IdleClient) {
		err := tPool.Put(ic)
		if err != nil {
			return
		}
	}(ic)
	res, err := ic.Client.(*raft.RaftServerClient).RequestVote(ctx, req)
	if err != nil {
		logger.WithContext(ctx).Errorf("Error during RequestVote:%v", err)
	}
	return res, err
}

var MapPool = thriftPool.NewMapPool(200, 1, 10, newClient, Close)

func Close(c *thriftPool.IdleClient) error {
	return c.Transport.Close()
}

func newClient(server string, connTimeout time.Duration) (*thriftPool.IdleClient, error) {
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
			return nil, err
		}
	}
	return &thriftPool.IdleClient{
		Transport: tr,
		Client:    client,
	}, err
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
	if err := transport.Open(); err != nil {
		return nil, nil, err
	}
	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)
	return raft.NewRaftServerClient(thrift.NewTStandardClient(iprot, oprot)), transport, nil
}
