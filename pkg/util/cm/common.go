package cm

import (
	"io"
	"os"
	"os/exec"
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

func Mkdir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}
}

func Rmdir(path string) {
	d, err := os.Stat(path)
	if err != nil {
		return
	}
	if d.IsDir() {
		os.RemoveAll(path)
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

func Call(param string) ([]byte, error) {
	c := exec.Command("/bin/bash", "-c", param)
	return c.CombinedOutput()
}

func Execute(param string) error {
	c := exec.Command("/bin/bash", "-c", param)
	return c.Run()
}
