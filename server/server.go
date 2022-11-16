package server

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
	"_9932xt/myraft/conf"
	"_9932xt/myraft/gen-go/raft"
	"_9932xt/myraft/handler"
	"_9932xt/myraft/heartbeat"
	log2 "_9932xt/myraft/log"
	"_9932xt/myraft/state"
	"_9932xt/myraft/util"
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"log"
	"os"
	"strconv"
	"strings"
)

var raftHandler = handler.NewRaftHandler()
var clientRaftHandler = handler.NewClientRaftHandler()

func RunServer(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, secure bool,
	id int, curConf *conf.Conf, cfg *thrift.TConfiguration) error {
	var transport thrift.TServerTransport
	var clientTransport thrift.TServerTransport
	state.GetServerState().ServerId = id
	state.GetServerState().Conf = curConf
	serverAddr, ok := curConf.InnerAddrMap[(id)]
	if !ok {
		log.Fatalf("invalid id:%d", id)
	}
	state.GetServerState().Net.ServerAddr = serverAddr
	clientAddr, ok := curConf.OuterAddrMap[id]
	if !ok {
		log.Fatalf("invalid id:%d", id)
	}
	state.GetServerState().Net.ClientAddr = clientAddr

	err := initPersistenceStatus(curConf, state.GetServerState().ServerId)
	if err != nil {
		log.Fatalf("initPersistenceStatus err:%v", err)
	}
	//go func() {
	//	registerPeerServer(transportFactory, protocolFactory, secure, cfg)
	//}()

	fmt.Printf("server state: %+v\n", state.GetServerState())

	if secure {
		cfg := new(tls.Config)
		if cert, err := tls.LoadX509KeyPair("server.crt", "server.key"); err == nil {
			cfg.Certificates = append(cfg.Certificates, cert)
		} else {
			return err
		}
		transport, err = thrift.NewTSSLServerSocket(serverAddr, cfg)
		clientTransport, err = thrift.NewTSSLServerSocket(clientAddr, cfg)
	} else {
		transport, err = thrift.NewTServerSocket(serverAddr)
		clientTransport, err = thrift.NewTServerSocket(clientAddr)
	}

	if err != nil {
		return err
	}
	fmt.Printf("%T\n", transport)
	processor := raft.NewRaftServerProcessor(raftHandler)
	clientProcessor := raft.NewClientRaftServerProcessor(clientRaftHandler)
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	clientServer := thrift.NewTSimpleServer4(clientProcessor, clientTransport, transportFactory, protocolFactory)

	go func() {
		Run()
	}()
	go func() {
		heartbeat.Run()
	}()
	go func() {
		fmt.Println("Starting myraft clientServer... on ", clientAddr)
		err := clientServer.Serve()
		if err != nil {
			log.Fatalf("client server err:%v", err)
		}
	}()
	fmt.Println("Starting myraft server... on ", serverAddr)
	return server.Serve()
}

func initPersistenceStatus(curconf *conf.Conf, serverId int) error {
	logDir := curconf.LogDir
	logFilePath := logDir + strconv.FormatInt(int64(serverId), 10) + ".log"
	if util.FileIsExist(logFilePath) {
		logFile, err := os.Open(logFilePath)
		if err != nil {
			log.Println(err)
			return err
		}
		defer func(logFile *os.File) {
			err := logFile.Close()
			if err != nil {
				log.Println(err)
			}
		}(logFile)
		scanner := bufio.NewScanner(logFile)
		logs := make([]*raft.LogEntry, 0)
		for scanner.Scan() {
			lineText := scanner.Text()
			slog, err := log2.ParseLog(lineText)
			if err != nil {
				return err
			}
			logs = append(logs, slog)
		}
		state.GetServerState().PersistentState.Logs = logs
	} else {
		err := util.CreateFile(logFilePath)
		if err != nil {
			return err
		}
	}

	dataFilePath := logDir + strconv.FormatInt(int64(serverId), 10) + ".data"
	if util.FileIsExist(dataFilePath) {
		dataFile, err := os.Open(dataFilePath)
		if err != nil {
			log.Println(err)
			return err
		}
		defer func(dataFile *os.File) {
			err := dataFile.Close()
			if err != nil {
				log.Println(err)
			}
		}(dataFile)
		scanner := bufio.NewScanner(dataFile)
		for scanner.Scan() {
			lineText := scanner.Text()
			arr := strings.Split(lineText, "=")
			if len(arr) != 2 {
				log.Printf("invalid config:%s\n", lineText)
			}
			if arr[0] == conf.Term {
				currentTerm, err := strconv.ParseInt(arr[1], 10, 64)
				if err != nil {
					return err
				}
				state.GetServerState().PersistentState.CurrentTerm = currentTerm
			} else if arr[0] == conf.VotedFor {
				votedForArr := strings.Split(arr[1], "_")
				term, err := strconv.ParseInt(votedForArr[0], 10, 64)
				if err != nil {
					return err
				}
				candidate, err := strconv.ParseInt(votedForArr[1], 10, 64)
				if err != nil {
					return err
				}
				state.GetServerState().PersistentState.VotedFor = &state.VotedFor{
					Term:      term,
					Candidate: int(candidate),
				}
			} else {
				log.Println("invalid config:", lineText)
			}
		}
	} else {
		err := util.CreateFile(dataFilePath)
		if err != nil {
			return err
		}
	}
	return nil
}

//func registerPeerServer(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, secure bool, cfg *thrift.TConfiguration) {
//	for {
//		if len(state.GetServerState().Conf.InnerAddrMap) > len(state.GetServerState().Net.PeerServerMap) {
//			// 有client没有连上来
//			peerServerMap := make(map[int]*net.PeerServer)
//			for id, addr := range state.GetServerState().Conf.InnerAddrMap {
//				if addr != state.GetServerState().Net.ServerAddr {
//
//					_, err := rpc.NewRaftServerClient(addr)
//					if err != nil {
//						time.Sleep(10 * time.Millisecond)
//						continue
//					}
//					_, ok := peerServerMap[(id)]
//					if !ok {
//						peerServerMap[(id)] = &net.PeerServer{
//							ServerAddr: addr,
//						}
//					}
//				}
//			}
//			state.GetServerState().Net.PeerServerMap = peerServerMap
//		}
//		time.Sleep(500 * time.Millisecond)
//	}
//}
