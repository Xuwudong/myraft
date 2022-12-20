package main

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
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/Xuwudong/myraft/gen-go/raft"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/pool"
	"github.com/apache/thrift/lib/go/thrift"
	log "github.com/sirupsen/logrus"
	"os"
)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	flag.Usage = Usage
	id := flag.Int64("id", 1, "server_id")
	serverAddr := flag.String("server_addr", "localhost:8080", "server_port")
	clientAddr := flag.String("client_addr", "localhost:9090", "client_port")

	wReq := &raft.DoCommandReq{
		Command: &raft.Command{
			Opt: raft.Opt_Write,
			Entry: &raft.Entry{
				EntryType: raft.EntryType_MemberChange,
				AddMembers: []*raft.Member{
					{
						MemberID:   *id,
						ServerAddr: *serverAddr,
						ClientAddr: *clientAddr,
					},
				},
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
	log.Print(resp)
}
