package rpc

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/net"
	"github.com/apache/thrift/lib/go/thrift"
	"log"
)

func NewClient(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string,
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
	//defer transport.Close()
	if err := transport.Open(); err != nil {
		return nil, nil, err
	}
	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)
	return raft.NewClientRaftServerClient(thrift.NewTStandardClient(iprot, oprot)), transport, nil
}

func NewRaftServerClient(addr string) (*raft.RaftServerClient, error) {
	c, transport, err := NewClient(net.TransportFactory, net.ProtocolFactory, addr, net.Secure, net.Cfg)
	if err != nil {
		log.Printf("error new client: %v", err)
		return c, err
	}
	defer func(transport thrift.TTransport) {
		err := transport.Close()
		if err != nil {
			log.Printf("transsport error:%v", err)
		}
	}(transport)
	return c, nil
}
