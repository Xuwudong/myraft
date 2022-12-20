package conf

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/Xuwudong/myraft/logger"
)

const LogDor string = "logDir"
const Term string = "term"
const VotedFor string = "votedFor"

type Conf struct {
	LogDir string
}

func ParseConf() (*Conf, error) {
	file, err := os.Open("conf/raft.conf")
	if err != nil {
		logger.Error(err)
	}
	conf := &Conf{}
	defer file.Close()
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
		} else {
			logger.Errorf("invalid conf:%s\n", lineText)
		}
	}
	logger.WithContext(context.Background()).Infof("load conf:%+v", conf)
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
			logger.Errorf("invalid conf:%s", line)
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
