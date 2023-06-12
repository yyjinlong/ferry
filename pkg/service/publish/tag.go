// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/config"
	"nautilus/pkg/model"
)

func NewBuildTag(pid int64, serviceName string) error {
	pidStr := strconv.FormatInt(pid, 10)
	serviceObj, err := model.GetServiceInfo(serviceName)
	if err != nil {
		return fmt.Errorf(config.DB_QUERY_SERVICE_ERROR, serviceName, err)
	}

	if serviceObj.Lock != "" && serviceObj.Lock != pidStr {
		return fmt.Errorf(config.TAG_OPERATE_FORBIDDEN, pidStr)
	}

	if err := model.SetLock(serviceObj.ID, pidStr); err != nil {
		return fmt.Errorf(config.DB_WRITE_LOCK_ERROR, pidStr, err)
	}

	updateList, err := model.FindUpdateInfo(pid)
	if err != nil {
		return fmt.Errorf(config.TAG_QUERY_UPDATE_ERROR, err)
	}

	_, curPath, _, _ := runtime.Caller(1)
	var (
		mainPath   = filepath.Dir(filepath.Dir(filepath.Dir(curPath)))
		scriptPath = filepath.Join(mainPath, "script")
	)

	for _, item := range updateList {
		branch := item.DeployBranch
		codeModule, err := model.GetCodeModuleInfo(item.CodeModule)
		if err != nil {
			return fmt.Errorf(config.TAG_QUERY_UPDATE_ERROR, err)
		}
		lang := codeModule.Language
		addr := codeModule.RepoAddr
		module := codeModule.Name

		output := ""
		param := fmt.Sprintf("%s/maketag -s %s -m %s -l %s -a %s -b %s -i %d", scriptPath, serviceName, module, lang, addr, branch, pid)
		log.Infof("maketag command: %s", param)
		if err := CallRealtimeOut(param, &output, nil); err != nil {
			return fmt.Errorf(config.TAG_BUILD_FAILED, err)
		}
		fmt.Println("输出: ", output)
	}
	return nil
}

func NewReceiveTag(pid int64, module, tag string) error {
	log.Infof("receive module: %s build tag value: %s", module, tag)
	if err := model.UpdateTag(pid, module, tag); err != nil {
		return fmt.Errorf(config.TAG_UPDATE_DB_ERROR, err)
	}
	log.Infof("module: %s update tag: %s success", module, tag)
	return nil
}

func NewReceivePkg(pid int64, module, pkg string) error {
	log.Infof("receive module: %s compile package: %s", module, pkg)
	if err := model.UpdatePkg(pid, module, pkg); err != nil {
		return fmt.Errorf(config.PKG_UPDATE_DB_ERROR, err)
	}
	log.Infof("module: %s update pkg: %s success", module, pkg)
	return nil
}
