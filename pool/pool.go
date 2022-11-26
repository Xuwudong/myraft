package pool

import (
	"crypto/tls"
	"fmt"
	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	"github.com/apache/thrift/lib/go/thrift"
	"sync"
)

type Client struct {
	lock   sync.Mutex
	Client *raft.RaftServerClient
	inUse  bool
	tr     thrift.TTransport
	//invalid bool
	server string
}

var clientPool = make(map[string][]*Client, 0)

func Init(servers []string) {
	for _, addr := range servers {
		clientPool[addr] = make([]*Client, 0)
		for i := 0; i < 10; i++ {
			client, tr, err := newClient(addr)
			if err != nil {
				logger.Errorf("new Client error:%v", err)
				continue
			}
			clientPool[addr] = append(clientPool[addr], &Client{
				Client: client,
				tr:     tr,
				server: addr,
			})
		}
	}
}

func GetClientByServer(server string) (*Client, error) {
	clients, ok := clientPool[server]
	if !ok {
		return nil, fmt.Errorf("invalid server")
	}
	for _, client := range clients {
		if getClient(client) {
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
		Client: client,
		tr:     tf,
		server: server,
	}
	clients = append(clients, c)
	return c, nil
}

func Return(client *Client) error {
	client.lock.Lock()
	defer client.lock.Unlock()
	client.inUse = false
	return nil
}

func Recycle(client *Client) error {
	client.lock.Lock()
	defer client.lock.Unlock()
	client.inUse = false
	err := client.tr.Close()
	if err != nil {
		logger.Errorf("close client error:%v", err)
		return err
	}
	c, tr, err := newClient(client.server)
	if err != nil {
		logger.Errorf("new Client error:%v", err)
		return err
	}
	client.Client = c
	client.tr = tr
	return nil
}

func getClient(client *Client) bool {
	client.lock.Lock()
	defer client.lock.Unlock()
	if !client.inUse {
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

//func poolRun() {
//	for {
//		for server, clients := range clientPool {
//			if len(clients) < 10 {
//				for i := len(clients); i < 10; i++ {
//					client, tr, err := newClient(server)
//					if err != nil {
//						logger.Errorf("new Client error:%v", err)
//						continue
//					}
//					clientPool[server] = append(clientPool[server], &Client{
//						Client: client,
//						tr:     tr,
//					})
//					logger.Debugf("append client success, server:%s", server)
//				}
//			} else {
//				for _, c := range clients {
//					update, err := checkAndUpdate(server, c)
//					if err != nil {
//						logger.Errorf("checkAndUpdate Client err:%v", err)
//						continue
//					}
//					if update {
//						logger.Infof("update client success, server:%s", server)
//					}
//				}
//			}
//		}
//		//logger.WithContext(ctx).Printf("pool run")
//		time.Sleep(10 * time.Millisecond)
//	}
//}

//func checkAndUpdate(server string, client *Client) (bool, error) {
//	client.lock.Lock()
//	defer client.lock.Unlock()
//	if client.invalid {
//		c, tr, err := newClient(server)
//		if err != nil {
//			logger.Errorf("new Client error:%v", err)
//			return false, err
//		}
//		logger.Info("update Client success")
//		client.Client = c
//		client.tr = tr
//		client.invalid = false
//		return true, nil
//	}
//	return false, nil
//}

func NewClientServerClient(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string,
	secure bool, cfg *thrift.TConfiguration) (*raft.ClientRaftServerClient, thrift.TTransport, error) {
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
	return raft.NewClientRaftServerClient(thrift.NewTStandardClient(iprot, oprot)), transport, nil
}
