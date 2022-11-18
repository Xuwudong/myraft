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
	conf2 "_9932xt/myraft/conf"
	"_9932xt/myraft/net"
	server2 "_9932xt/myraft/server"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/apache/thrift/lib/go/thrift"
)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	flag.Usage = Usage
	server := flag.Bool("server", true, "Run server")
	protocol := flag.String("P", "binary", "Specify the protocol (binary, compact, json, simplejson)")
	framed := flag.Bool("framed", false, "Use framed transport")
	buffered := flag.Bool("buffered", false, "Use buffered transport")
	secure := flag.Bool("secure", false, "Use tls secure transport")
	id := flag.Int("id", 1, "server_id")

	flag.Parse()

	initLog(*id)

	switch *protocol {
	case "compact":
		net.ProtocolFactory = thrift.NewTCompactProtocolFactoryConf(nil)
	case "simplejson":
		net.ProtocolFactory = thrift.NewTSimpleJSONProtocolFactoryConf(nil)
	case "json":
		net.ProtocolFactory = thrift.NewTJSONProtocolFactory()
	case "binary", "":
		net.ProtocolFactory = thrift.NewTBinaryProtocolFactoryConf(nil)
	default:
		fmt.Fprint(os.Stderr, "Invalid protocol specified", protocol, "\n")
		Usage()
		os.Exit(1)
	}

	net.Cfg = &thrift.TConfiguration{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	if *buffered {
		net.TransportFactory = thrift.NewTBufferedTransportFactory(8192)
	} else {
		net.TransportFactory = thrift.NewTTransportFactory()
	}

	if *framed {
		net.TransportFactory = thrift.NewTFramedTransportFactoryConf(net.TransportFactory, net.Cfg)
	}

	conf, err := conf2.ParseConf()
	if err != nil {
		log.Fatal(err)
	}
	net.Secure = *secure
	if *server {
		if err := server2.RunServer(net.TransportFactory, net.ProtocolFactory, *secure, *id, conf, net.Cfg); err != nil {
			fmt.Println("error running server:", err)
		}
	}
}

func initLog(id int) {
	f, err := os.OpenFile("log"+strconv.Itoa(id)+".log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		return
	}
	//defer func() {
	//	f.Close()
	//}()

	// 组合一下即可，os.Stdout代表标准输出流
	multiWriter := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multiWriter)

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("start server:%d \n", id)
}
