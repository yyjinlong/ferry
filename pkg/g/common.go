// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package g

import (
	"io"
	"os"
	"time"

	"github.com/satori/go.uuid"
)

func In(data string, dataList []string) bool {
	for _, item := range dataList {
		if data == item {
			return true
		}
	}
	return false
}

func Ini(num int, numList []int) bool {
	for _, n := range numList {
		if num == n {
			return true
		}
	}
	return false
}

func TimeString(curTime time.Time) string {
	return curTime.Format("2006-01-02 15:04:05")
}

func UniqueID() string {
	return uuid.NewV4().String()
}

func Mkdir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}
}

func Copy(source, dest string) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
}
