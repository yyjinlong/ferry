// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package git

import (
	"os"
	"path/filepath"

	"nautilus/pkg/g"
	"nautilus/pkg/log"
)

func DownloadCode(module, repo, tag, codePath string) {
	mygit := &MyGit{
		module:   module,
		repo:     repo,
		tag:      tag,
		codePath: codePath,
	}

	if err := mygit.Clone(); err != nil {
		log.Errorf("git clone module: %s failed: %s", module, err)
		return
	}
	log.Infof("git clone module: %s success", module)

	if err := mygit.CheckoutTag(); err != nil {
		log.Errorf("git checkout tag: %s failed: %s", tag, err)
		return
	}
	log.Infof("git checkout tag: %s success", tag)
}

type MyGit struct {
	module   string
	repo     string
	tag      string
	codePath string
}

func (mg *MyGit) Clone() error {
	directory := filepath.Join(mg.codePath, mg.module)
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if _, err := g.Cmd("git", "clone", mg.repo, directory); err != nil {
			return err
		}
	}
	return nil
}

func (mg *MyGit) CheckoutTag() error {
	directory := filepath.Join(mg.codePath, mg.module)
	if _, err := g.Cmd2(directory, "git", "checkout", mg.tag); err != nil {
		return err
	}
	return nil
}
