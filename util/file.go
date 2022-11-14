package util

import (
	"os"
)

func FileIsExist(file string) bool {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func CreateFile(file string) error {
	f, err := os.Create(file)
	defer f.Close()
	return err
}
