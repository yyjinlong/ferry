// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package g

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"

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

func Cmd(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	return _do(cmd)
}

func Cmd2(dir string, name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	return _do(cmd)
}

func _do(cmd *exec.Cmd) ([]byte, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	output, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return output, nil
}

func Execute(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	return cmd.Run()
}

func ExecuteByDir(dir string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	return cmd.Run()
}
