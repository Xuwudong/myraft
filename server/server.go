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
	log2 "_9932xt/myraft/log"
	"_9932xt/myraft/net"
	"_9932xt/myraft/rpc"
	"_9932xt/myraft/util"
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
)

var server = &Server{
	Role:                raft.Role_Follower,
	Net:                 &net.Net{},
	PersistentState:     &PersistentState{},
	VolatileState:       &VolatileState{},
	MasterVolatileState: &MasterVolatileState{},
}

type Server struct {
	ServerId            int
	Role                raft.Role
	Net                 *net.Net
	PersistentState     *PersistentState
	VolatileState       *VolatileState
	MasterVolatileState *MasterVolatileState
	Conf                *conf.Conf
}

type PersistentState struct {
	// 服务器已知最新的任期（在服务器首次启动时初始化为0，单调递增）
	CurrentTerm int64
	// 当前任期内收到选票的 candidateId，如果没有投给任何候选人 则为空
	VotedFor int
	// 日志条目；每个条目包含了用于状态机的命令，以及领导人接收到该条目时的任期（初始索引为1）
	Logs []*log2.Log
}

type VolatileState struct {
	// 已知已提交的最高的日志条目的索引（初始值为0，单调递增）
	CommitIndex int64
	// 已经被应用到状态机的最高的日志条目的索引（初始值为0，单调递增）
	LastApplied int64
	// 选票数
	VoteNum uint32
}

type MasterVolatileState struct {
	// 对于每一台服务器，发送到该服务器的下一个日志条目的索引（初始值为领导人最后的日志条目的索引+1）
	nextIndexMap map[string]int64
	// 对于每一台服务器，已知的已经复制到该服务器的最高日志条目的索引（初始值为0，单调递增）
	matchIndexMap map[string]int64
}

func RunServer(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, secure bool,
	id int, curconf *conf.Conf, cfg *thrift.TConfiguration) error {
	var transport thrift.TServerTransport
	var clientTransport thrift.TServerTransport
	server.ServerId = id
	server.Conf = curconf
	if server.Net == nil {
		server.Net = &net.Net{}
	}
	serverAddr, ok := curconf.InnerAddrMap[conf.Server+"."+strconv.FormatInt(int64(id), 10)]
	if !ok {
		log.Fatalf("invalid id:%d", id)
	}
	server.Net.ServerAddr = serverAddr
	go func() {
		registerPeerServer(transportFactory, protocolFactory, secure, cfg)
	}()
	clientAddr, ok := curconf.OuterAddrMap[conf.Server+"."+strconv.FormatInt(int64(id), 10)]
	if !ok {
		log.Fatalf("invalid id:%d", id)
	}
	server.Net.ClientAddr = clientAddr

	err := initPersistenceStatus(curconf, server.ServerId)
	if err != nil {
		log.Fatalf("initPersistenceStatus err:%v", err)
	}

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
	voteHandler := handler.NewRaftHandler()
	processor := raft.NewRaftServerProcessor(voteHandler)
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	clientServer := thrift.NewTSimpleServer4(processor, clientTransport, transportFactory, protocolFactory)

	go func() {
		Run()
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

func GetServer() *Server {
	return server
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
		logs := make([]*log2.Log, 0)
		for scanner.Scan() {
			lineText := scanner.Text()
			slog, err := log2.ParseLog(lineText)
			if err != nil {
				return err
			}
			logs = append(logs, slog)
		}
		GetServer().PersistentState.Logs = logs
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
		lineText := scanner.Text()
		if lineText != "" {
			currentTerm, err := strconv.ParseInt(lineText, 10, 64)
			if err != nil {
				return err
			}
			GetServer().PersistentState.CurrentTerm = currentTerm
		}
		lineText = scanner.Text()
		if lineText != "" {
			votedFor, err := strconv.ParseInt(lineText, 10, 64)
			if err != nil {
				return err
			}
			GetServer().PersistentState.VotedFor = int(votedFor)
		}
	} else {
		err := util.CreateFile(dataFilePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func registerPeerServer(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, secure bool, cfg *thrift.TConfiguration) {
	for {
		if len(GetServer().Conf.InnerAddrMap) > len(GetServer().Net.PeerServerMap) {
			// 有client没有连上来
			peerServerMap := make(map[int]*net.PeerServer)
			for serverId, addr := range GetServer().Conf.InnerAddrMap {
				if addr != GetServer().Net.ServerAddr {
					idStr := strings.Replace(serverId, conf.Server, "", -1)
					idd, _ := strconv.ParseInt(idStr, 10, 32)

					client, err := rpc.NewClient(transportFactory, protocolFactory, addr, secure, cfg)
					if err != nil {
						log.Printf("error new client: %v", err)
						continue
					}
					_, ok := peerServerMap[int(idd)]
					if !ok {
						peerServerMap[int(idd)] = &net.PeerServer{
							ServerAddr: addr,
							Client:     client,
						}
					}
				}
			}
			GetServer().Net.PeerServerMap = peerServerMap
		}
		time.Sleep(100 * time.Millisecond)
	}
}
