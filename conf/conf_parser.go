package conf

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const LogDor string = "logDir"
const Server string = "server"

type Conf struct {
	LogDir       string
	InnerAddrMap map[string]string
	OuterAddrMap map[string]string
}

func ParseConf() (*Conf, error) {
	file, err := os.Open("conf/raft.conf")
	if err != nil {
		log.Println(err)
	}
	conf := &Conf{}
	defer file.Close()
	innerAddr := make(map[string]string, 0)
	outerAddr := make(map[string]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineText := scanner.Text()
		arr := strings.Split(lineText, "=")
		if len(arr) != 2 {
			return conf, fmt.Errorf("conf error:%s", lineText)
		}
		key := strings.Trim(arr[0], " ")
		if LogDor == key {
			conf.LogDir = strings.Trim(arr[1], " ")
		} else if strings.Index(key, Server) >= 0 {
			value := strings.Trim(arr[1], " ")
			addrArr := strings.Split(value, ":")

			innerAddr[key] = addrArr[0] + ":" + addrArr[1]
			outerAddr[key] = addrArr[0] + ":" + addrArr[2]
		} else {
			log.Println("invalid conf")
		}
	}
	conf.InnerAddrMap = innerAddr
	conf.OuterAddrMap = outerAddr
	log.Printf("load conf:%+v \n", conf)
	return conf, nil

}
