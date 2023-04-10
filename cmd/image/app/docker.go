// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package app

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/model"
	"nautilus/pkg/util/cm"
)

// Dockerfile generate docker file on build dir
func Dockerfile(buildPath string) error {
	tpl := `
ARG repo
FROM ${repo}

ARG deploy_path=/home/tong/www

ADD --chown=tong:tong ./code ${deploy_path}`

	dockerfile := filepath.Join(buildPath, "Dockerfile")
	if err := ioutil.WriteFile(dockerfile, []byte(tpl), 0644); err != nil {
		log.Errorf("generate dockerfile: %s error: %+v", dockerfile, err)
		return err
	}
	log.Infof("generate dockerfile: %s success", dockerfile)
	return nil
}

// DockerBuild 基于服务镜像, 构建包含代码的release镜像
func DockerBuild(service, targetURL, buildPath string) error {
	si, err := model.GetServiceInfo(service)
	if err != nil {
		log.Errorf("query service: %s failed: %s", service, err)
		return err
	}

	var (
		serviceImageAddr = si.ImageAddr
		pull             = fmt.Sprintf("docker pull %s", serviceImageAddr)
		repo             = fmt.Sprintf("repo=%s", serviceImageAddr)
		param            = fmt.Sprintf("docker build --build-arg %s -t %s %s", repo, targetURL, buildPath)
	)

	// docker pull 服务镜像
	log.Info(pull)
	if err := cm.Execute(pull); err != nil {
		log.Errorf("docker pull error: %+v", err)
		return err
	}
	log.Infof("docker pull service image: %s success", serviceImageAddr)

	// docker build --build-arg repo=服务镜像 -t release镜像:版本 dockerfile路径
	log.Info(param)
	if err := cm.Execute(param); err != nil {
		log.Errorf("docker build error: %+v", err)
		return err
	}
	log.Infof("docker build release image: %s success", targetURL)
	return nil
}

// DockerPush docker push
func DockerPush(targetURL string) error {
	param := fmt.Sprintf("docker push %s", targetURL)
	log.Info(param)
	if err := cm.Execute(param); err != nil {
		log.Errorf("docker push release image: %s error: %+v", targetURL, err)
		return err
	}
	log.Infof("docker push release image: %s success", targetURL)
	return nil
}
