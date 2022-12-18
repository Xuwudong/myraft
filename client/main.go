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
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"

	conf2 "github.com/Xuwudong/myraft/conf"
	"github.com/Xuwudong/myraft/logger"
	"github.com/Xuwudong/myraft/net"
	server2 "github.com/Xuwudong/myraft/server"
	log "github.com/sirupsen/logrus"

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
	serverPort := flag.Int64("serverPort", 8080, "server_port")
	clientPort := flag.Int64("clientPort", 9090, "client_port")

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
		logger.Fatal(err)
	}
	net.Secure = *secure
	if *server {
		if err := server2.RunServer(net.TransportFactory, net.ProtocolFactory, *secure, *id, conf, *serverPort, *clientPort); err != nil {
			fmt.Println("error running server:", err)
		}
	}
}

func initLog(id int) {
	// 设置日志格式为json格式
	//logger.WithContext(ctx).SetFormatter(&logger.WithContext(ctx).JSONFormatter{})
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05.000000"
	customFormatter.FullTimestamp = true
	customFormatter.CallerPrettyfier = func(frame *runtime.Frame) (function string, file string) {
		fileName := path.Base(frame.File) + ":" + strconv.Itoa(frame.Line)
		//return frame.Function, fileName
		return "", fileName
	}
	log.SetFormatter(customFormatter)

	// 设置日志级别为warn以上
	//logger.WithContext(ctx).SetLevel(logger.WithContext(ctx).WarnLevel)

	f, err := os.OpenFile("run_log"+strconv.Itoa(id)+".log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		return
	}
	multiWriter := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multiWriter)
	log.SetReportCaller(true)

	log.Printf("start server:%d \n", id)
}
