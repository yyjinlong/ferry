// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package g

import (
	"os"
	"path/filepath"
)

func NewGit(module, repo, tag, codePath string) *git {
	return &git{
		module:   module,
		repo:     repo,
		tag:      tag,
		codePath: codePath,
	}
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
