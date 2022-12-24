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
	"bufio"
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Xuwudong/myraft/conf"
	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/handler"
	"github.com/Xuwudong/myraft/heartbeat"
	log2 "github.com/Xuwudong/myraft/log"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/middleware"
	"github.com/Xuwudong/myraft/state"
	"github.com/Xuwudong/myraft/util"
	"github.com/apache/thrift/lib/go/thrift"
)

var raftHandler = handler.NewRaftHandler()
var clientRaftHandler = handler.NewClientRaftHandler()

func RunServer(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, secure bool,
	id int, curConf *conf.Conf, serverPort, clientPort int64) error {
	var transport thrift.TServerTransport
	var clientTransport thrift.TServerTransport
	state.GetServerState().ServerId = id
	state.GetServerState().Conf = curConf
	state.GetServerState().Net.ServerAddr = "localhost:" + strconv.FormatInt(serverPort, 10)
	state.GetServerState().Net.ClientAddr = "localhost:" + strconv.FormatInt(clientPort, 10)

	err := initPersistenceStatus(curConf, state.GetServerState().ServerId)
	if err != nil {
		logger.Fatalf("initPersistenceStatus err:%v", err)
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
		transport, err = thrift.NewTSSLServerSocket(state.GetServerState().Net.ServerAddr, cfg)
		clientTransport, err = thrift.NewTSSLServerSocket(state.GetServerState().Net.ClientAddr, cfg)
	} else {
		transport, err = thrift.NewTServerSocket(state.GetServerState().Net.ServerAddr)
		clientTransport, err = thrift.NewTServerSocket(state.GetServerState().Net.ClientAddr)
	}

	if err != nil {
		return err
	}
	fmt.Printf("%T\n", transport)
	processor := raft.NewRaftServerProcessor(raftHandler)
	wrapper := thrift.WrapProcessor(processor, middleware.TrackIdProcessorMiddleware())
	server := thrift.NewTSimpleServer4(wrapper, transport, transportFactory, protocolFactory)

	clientProcessor := raft.NewClientRaftServerProcessor(clientRaftHandler)
	cwrapper := thrift.WrapProcessor(clientProcessor, middleware.TrackIdProcessorMiddleware())
	clientServer := thrift.NewTSimpleServer4(cwrapper, clientTransport, transportFactory, protocolFactory)

	state.GetServerState().MemberConf.ServerAddrMap[int64(id)] = state.GetServerState().Net.ServerAddr
	state.GetServerState().MemberConf.ClientAddrMap[int64(id)] = state.GetServerState().Net.ClientAddr
	//pool.Init(state.GetServerState().VolatileState.PeerServers)
	go func() {
		Run()
	}()
	go func() {
		ToNewServerAddrMapLoop()
	}()
	go func() {
		heartbeat.Run()
	}()
	go func() {
		logger.Infof("Starting myraft clientServer... on ", state.GetServerState().Net.ClientAddr)
		err := clientServer.Serve()
		if err != nil {
			logger.Fatalf("client server err:%v", err)
		}
	}()
	logger.Infof("Starting myraft server... on ", state.GetServerState().Net.ServerAddr)
	return server.Serve()
}

func initPersistenceStatus(curconf *conf.Conf, serverId int) error {
	logDir := curconf.LogDir
	logFilePath := logDir + strconv.FormatInt(int64(serverId), 10) + ".log"
	if util.FileIsExist(logFilePath) {
		logFile, err := os.Open(logFilePath)
		if err != nil {
			logger.Error(err)
			return err
		}
		defer func(logFile *os.File) {
			err := logFile.Close()
			if err != nil {
				logger.Error(err)
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
			logger.Error(err)
			return err
		}
		defer func(dataFile *os.File) {
			err := dataFile.Close()
			if err != nil {
				logger.Error(err)
			}
		}(dataFile)
		scanner := bufio.NewScanner(dataFile)
		for scanner.Scan() {
			lineText := scanner.Text()
			arr := strings.Split(lineText, "=")
			if len(arr) != 2 {
				logger.Errorf("invalid config:%s\n", lineText)
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
				logger.Error("invalid config:", lineText)
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
