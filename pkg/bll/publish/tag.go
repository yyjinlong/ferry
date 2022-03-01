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

	"nautilus/internal/objects"
	"nautilus/pkg/base"
	"nautilus/pkg/log"
)

type BuildTag struct{}

func (bt *BuildTag) Handle(r *base.Request) (interface{}, error) {
	type params struct {
		ID      int64  `form:"pipeline_id" binding:"required"`
		Service string `form:"service" binding:"required"`
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		return "", err
	}

	var (
		pid         = data.ID
		pidStr      = strconv.FormatInt(pid, 10)
		serviceName = data.Service
	)
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid})

	serviceObj, err := objects.GetServiceInfo(serviceName)
	if err != nil {
		return "", fmt.Errorf(DB_QUERY_SERVICE_ERROR, serviceName, err)
	}

	if serviceObj.Lock != "" && serviceObj.Lock != pidStr {
		return "", fmt.Errorf(TAG_OPERATE_FORBIDDEN, pidStr)
	}

	if err := objects.SetLock(serviceObj.Service.ID, pidStr); err != nil {
		return "", fmt.Errorf(TAG_WRITE_LOCK_ERROR, pidStr, err)
	}

	updateList, err := objects.FindUpdateInfo(pid)
	if err != nil {
		log.Errorf("find pipeline update info error: %s", err)
		return nil, fmt.Errorf(TAG_QUERY_UPDATE_ERROR, err)
	}

	_, curPath, _, _ := runtime.Caller(1)
	var (
		mainPath   = filepath.Dir(filepath.Dir(filepath.Dir(curPath)))
		scriptPath = filepath.Join(mainPath, "script")
	)

	for _, item := range updateList {
		addr := item.CodeModule.ReposAddr
		module := item.CodeModule.Name
		branch := item.PipelineUpdate.DeployBranch
		param := fmt.Sprintf("%s/maketag -a %s -m %s -b %s -i %d", scriptPath, addr, module, branch, pid)
		log.Infof("maketag command: %s", param)
		if !bt.do(param) {
			return "", fmt.Errorf(TAG_BUILD_FAILED)
		}
	}
	return "", nil
}

func (bt *BuildTag) do(param string) bool {
	cmd := exec.Command("/bin/bash", "-c", param)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(TAG_CREATE_PIPE_ERROR, err)
		return false
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		fmt.Println(TAG_START_EXEC_ERROR, err)
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
		fmt.Println(TAG_WAIT_FINISH_ERROR, err)
		return false
	}

	if cmd.ProcessState.Success() {
		return true
	}
	return false
}

type ReceiveTag struct{}

func (rt *ReceiveTag) Handle(r *base.Request) (interface{}, error) {
	type params struct {
		ID     int64  `form:"taskid" binding:"required"`
		Module string `form:"module" binding:"required"`
		Tag    string `form:"tag" binding:"required"`
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		return "", err
	}

	var (
		pid    = data.ID
		module = data.Module
		tag    = data.Tag
	)
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid})
	log.Infof("receive module: %s build tag value: %s", module, tag)

	if err := objects.UpdateTag(pid, module, tag); err != nil {
		return "", fmt.Errorf(TAG_UPDATE_DB_ERROR, err)
	}
	log.Infof("module: %s update tag: %s success", module, tag)
	return "", nil
}
