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
	"nautilus/pkg/util/cm"
)

// DockerfileCopy copy docker file to build dir
func DockerfileCopy(appPath, buildPath string) error {
	var (
		srcFile = filepath.Join(appPath, "app", "Dockerfile")
		dstFile = filepath.Join(buildPath, "Dockerfile")
	)

	if err := cm.Copy(srcFile, dstFile); err != nil {
		log.Errorf("copy dockerfile: %s failed: %s", srcFile, err)
		return err
	}
	log.Infof("copy dockerfile: %s success", dstFile)
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
