// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package app

import (
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/model"
	"nautilus/pkg/util"
)

type pipeline struct {
	pid       int64
	service   string
	appPath   string
	buildPath string
	codePath  string
	imageURL  string
	imageTag  string
	targetURL string
}

func (p *pipeline) copyDockerfile() bool {
	var (
		srcFile = filepath.Join(p.appPath, "app", "Dockerfile")
		dstFile = filepath.Join(p.buildPath, "Dockerfile")
	)
	if err := util.Copy(srcFile, dstFile); err != nil {
		log.Errorf("copy dockerfile: %s failed: %s", srcFile, err)
		return false
	}
	log.Infof("copy dockerfile: %s success", dstFile)
	return true
}

func (p *pipeline) dockerBuild() bool {
	// note: 基于服务镜像, 构建包含代码的release镜像
	si, err := model.GetServiceInfo(p.service)
	if err != nil {
		log.Errorf("query service: %s failed: %s", p.service, err)
		return false
	}

	var (
		serviceImageAddr = si.ImageAddr
		pull             = fmt.Sprintf("docker pull %s", serviceImageAddr)
		repo             = fmt.Sprintf("repo=%s", serviceImageAddr)
		param            = fmt.Sprintf("docker build --build-arg %s -t %s %s", repo, p.targetURL, p.buildPath)
	)

	// note: docker pull 服务镜像
	log.Info(pull)
	if err := util.Execute(pull); err != nil {
		log.Errorf("docker pull error: %+v", err)
		return false
	}
	log.Infof("docker pull service image: %s success", serviceImageAddr)

	// note: docker build --build-arg repo=服务镜像 -t release镜像:版本 dockerfile路径
	log.Info(param)
	if err := util.Execute(param); err != nil {
		log.Errorf("docker build error: %+v", err)
		return false
	}
	log.Infof("docker build release image: %s success", p.targetURL)
	return true
}

func (p *pipeline) dockerPush() bool {
	param := fmt.Sprintf("docker push %s", p.targetURL)
	log.Info(param)
	if err := util.Execute(param); err != nil {
		log.Errorf("docker push error: %+v", err)
		return false
	}
	log.Infof("docker push release image: %s success", p.targetURL)
	return true
}
