// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package g

import (
	"io/ioutil"
	"os/exec"
)

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
