// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/yyjinlong/golib/log"

	"nautilus/pkg/cfg"
	"nautilus/pkg/model"
)

func NewBuildTag() *BuildTag {
	return &BuildTag{}
}

type BuildTag struct{}

func (bt *BuildTag) Handle(pid int64, serviceName string) error {
	pidStr := strconv.FormatInt(pid, 10)
	serviceObj, err := model.GetServiceInfo(serviceName)
	if err != nil {
		return fmt.Errorf(cfg.DB_QUERY_SERVICE_ERROR, serviceName, err)
	}

	if serviceObj.Lock != "" && serviceObj.Lock != pidStr {
		return fmt.Errorf(cfg.TAG_OPERATE_FORBIDDEN, pidStr)
	}

	if err := model.SetLock(serviceObj.ID, pidStr); err != nil {
		return fmt.Errorf(cfg.TAG_WRITE_LOCK_ERROR, pidStr, err)
	}

	updateList, err := model.FindUpdateInfo(pid)
	if err != nil {
		return fmt.Errorf(cfg.TAG_QUERY_UPDATE_ERROR, err)
	}

	_, curPath, _, _ := runtime.Caller(1)
	var (
		mainPath   = filepath.Dir(filepath.Dir(filepath.Dir(curPath)))
		scriptPath = filepath.Join(mainPath, "script")
	)

	for _, item := range updateList {
		branch := item.DeployBranch
		codeModule, err := model.GetCodeModuleInfoByID(item.CodeModuleID)
		if err != nil {
			return fmt.Errorf(cfg.TAG_QUERY_UPDATE_ERROR, err)
		}
		addr := codeModule.ReposAddr
		module := codeModule.Name

		param := fmt.Sprintf("%s/maketag -a %s -m %s -b %s -i %d", scriptPath, addr, module, branch, pid)
		log.Infof("maketag command: %s", param)
		if !bt.do(param) {
			return fmt.Errorf(cfg.TAG_BUILD_FAILED)
		}
	}
	return nil
}

func (bt *BuildTag) do(param string) bool {
	cmd := exec.Command("/bin/bash", "-c", param)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(cfg.TAG_CREATE_PIPE_ERROR, err)
		return false
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		fmt.Println(cfg.TAG_START_EXEC_ERROR, err)
		return false
	}

	for {
		buf := make([]byte, 1024)
		_, err := stdout.Read(buf)
		fmt.Println(string(buf))
		if err != nil {
			break
		}
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println(cfg.TAG_WAIT_FINISH_ERROR, err)
		return false
	}

	if cmd.ProcessState.Success() {
		return true
	}
	return false
}

func NewReceiveTag() *ReceiveTag {
	return &ReceiveTag{}
}

type ReceiveTag struct{}

func (rt *ReceiveTag) Handle(pid int64, module, tag string) error {
	log.Infof("receive module: %s build tag value: %s", module, tag)
	if err := model.UpdateTag(pid, module, tag); err != nil {
		return fmt.Errorf(cfg.TAG_UPDATE_DB_ERROR, err)
	}
	log.Infof("module: %s update tag: %s success", module, tag)
	return nil
}
