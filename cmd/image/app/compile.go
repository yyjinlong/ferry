// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package app

import (
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/util/cm"
)

func DownloadCode(module, repo, tag, codePath string) error {
	directory := filepath.Join(codePath, module)
	cm.Rmdir(directory)
	log.Infof("download code directory: %s", directory)

	if err := cm.ExecuteDir(codePath, "git", "clone", repo); err != nil {
		return err
	}
	log.Infof("git clone %s success", repo)

	if err := cm.ExecuteDir(directory, "git", "checkout", tag); err != nil {
		return err
	}
	log.Infof("git checkout %s success", tag)
	return nil
}

func Compile(language string) error {
	switch language {
	case PYTHON:
	case GOLANG:
		// 执行Makefile
	}
	return nil
}
