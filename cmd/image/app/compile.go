// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package app

import (
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/util"
)

func downloadCode(module, repo, tag, codePath string) error {
	directory := filepath.Join(codePath, module)
	util.Rmdir(directory)
	log.Infof("download code directory: %s", directory)

	if err := util.ExecuteDir(codePath, "git", "clone", repo); err != nil {
		return err
	}
	log.Infof("git clone %s success", repo)

	if err := util.ExecuteDir(directory, "git", "checkout", tag); err != nil {
		return err
	}
	log.Infof("git checkout %s success", tag)
	return nil
}

func compile(language string) error {
	switch language {
	case PYTHON:
	case GOLANG:
	}
	return nil
}
