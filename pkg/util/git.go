// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package util

import (
	"path/filepath"

	"nautilus/golib/log"
)

func DownloadCode(module, repo, tag, codePath string) error {
	if err := Clone(codePath, module, repo); err != nil {
		return err
	}
	log.Infof("git clone module: %s success", module)

	if err := CheckoutTag(codePath, module, tag); err != nil {
		return err
	}
	log.Infof("git checkout tag: %s success", tag)
	return nil
}

func Clone(codePath, module, repo string) error {
	return ExecuteDir(codePath, "git", "clone", repo)
}

func CheckoutTag(codePath, module, tag string) error {
	directory := filepath.Join(codePath, module)
	return ExecuteDir(directory, "git", "checkout", tag)
}
