// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package g

import (
	"os"
	"path/filepath"

	"ferry/pkg/log"
)

func DownloadCode(module, repo, tag, codePath string) {
	mygit := &git{
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

type git struct {
	module   string
	repo     string
	tag      string
	codePath string
}

func (g *git) Clone() error {
	directory := filepath.Join(g.codePath, g.module)
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if _, err := Cmd("git", "clone", g.repo, directory); err != nil {
			return err
		}
	}
	return nil
}

func (g *git) CheckoutTag() error {
	directory := filepath.Join(g.codePath, g.module)
	if _, err := Cmd2(directory, "git", "checkout", g.tag); err != nil {
		return err
	}
	return nil
}
