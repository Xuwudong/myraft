package conf

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

const LogDor string = "logDir"
const Server string = "server"
const Term string = "term"
const VotedFor string = "votedFor"

type Conf struct {
	LogDir       string
	InnerAddrMap map[int]string
	OuterAddrMap map[int]string
}

func ParseConf() (*Conf, error) {
	file, err := os.Open("conf/raft.conf")
	if err != nil {
		log.Println(err)
	}
	conf := &Conf{}
	defer file.Close()
	innerAddr := make(map[int]string, 0)
	outerAddr := make(map[int]string, 0)
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
			idStr := strings.Replace(key, Server+".", "", -1)
			id, _ := strconv.ParseInt(idStr, 10, 64)

			innerAddr[int(id)] = addrArr[0] + ":" + addrArr[1]
			outerAddr[int(id)] = addrArr[0] + ":" + addrArr[2]
		} else {
			log.Printf("invalid conf:%s\n", lineText)
		}
	}
	conf.InnerAddrMap = innerAddr
	conf.OuterAddrMap = outerAddr
	log.Printf("load conf:%+v \n", conf)
	return conf, nil

}

func UpdateDataField(fileName string, field string, term, candidate int) error {
	lineBytes, err := ioutil.ReadFile(fileName)
	var lines []string
	if err != nil {
		fmt.Println(err)
	} else {
		contents := string(lineBytes)
		lines = strings.Split(contents, "\n")
	}
	var newLines []string

	flag := false
	for _, line := range lines {
		if line == "" {
			continue
		}
		arr := strings.Split(line, "=")
		if len(arr) != 2 {
			log.Printf("invalid conf:%s\n", line)
		}
		if arr[0] == field {
			if field == Term {
				arr[1] = strconv.Itoa(term)
				line = strings.Join(arr, "=")
				flag = true
			} else if field == VotedFor {
				arr[1] = strconv.Itoa(term) + "_" + strconv.Itoa(candidate)
				line = strings.Join(arr, "=")
				flag = true
			}
		}
		newLines = append(newLines, line)
	}
	if !flag {
		var line string
		if field == Term {
			line = field + "=" + strconv.Itoa(term)
		} else if field == VotedFor {
			line = field + "=" + strconv.Itoa(term) + "_" + strconv.Itoa(candidate)
		}
		//log.Printf("line:%s", line)
		newLines = append(newLines, line)
	}

	file, err := os.OpenFile(fileName, os.O_WRONLY, 0666)
	defer file.Close()
	_, err = file.WriteString(strings.Join(newLines, "\n"))
	if err != nil {
		return err
	}
	return nil
}
