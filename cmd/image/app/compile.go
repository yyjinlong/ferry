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

const (
	PYTHON   = "python"
	GOLANG   = "golang"
	JAVA_JAR = "jar"
	JAVA_MVN = "maven"
)

func Compile(module, repo, tag, codePath, language string) error {
	// (1) download code
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

	// (2) compile
	switch language {
	case GOLANG:
		// 执行Makefile, 编译成二进制

	case JAVA_JAR:
		// jar包格式: 执行Makefile, 调用maven编译

	case JAVA_MVN:
		// java程序: 执行Makefile, 调用maven编译

	default:
		// 默认动态类型语言, 不需要编译

	}
	return nil
}
