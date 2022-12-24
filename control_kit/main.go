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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Xuwudong/myraft/gen-go/raft"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	Init([]string{leader})
	http.HandleFunc("/member_add", memberAdd)

	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func memberAdd(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("could not read body: %s\n", err)
	}

	addMembers := make([]*raft.Member, 0)
	json.Unmarshal(body, &addMembers)
	log.Printf("addMembers:%v", addMembers)
	tempMembers := members[:]

	tempMembers = append(tempMembers, addMembers...)
	wReq := &raft.DoCommandReq{
		Command: &raft.Command{
			Opt: raft.Opt_Write,
			Entry: &raft.Entry{
				EntryType: raft.EntryType_MemberChange,
				Members:   tempMembers,
			},
		},
	}
	resp, err := DoCommand(context.Background(), wReq)
	log.Print(resp)
	if err == nil && resp.Succuess {
		members = tempMembers
	} else {
		log.Errorf("do command failed, req:%v", wReq)
	}
	log.Printf("members:%v", members)
	io.WriteString(w, fmt.Sprintf("resp:%v", resp))
}
