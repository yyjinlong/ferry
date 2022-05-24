// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package app

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"nautilus/golib/log"
	"nautilus/pkg/config"
	"nautilus/pkg/model"
	"nautilus/pkg/util"
)

func getCurPath() string {
	_, curPath, _, _ := runtime.Caller(1)
	return curPath
}

func getTag() string {
	return fmt.Sprintf("v-%s", time.Now().Format("20060102_150405"))
}

func worker(data Image) {
	var (
		pid       = data.PID
		service   = data.Service
		buildPath = filepath.Join(config.Config().Image.Dir, service, strconv.FormatInt(pid, 10))
		appPath   = filepath.Dir(filepath.Dir(getCurPath()))
		codePath  = filepath.Join(buildPath, "code")
		imageURL  = fmt.Sprintf("%s/%s", config.Config().Image.Registry, service)
		imageTag  = getTag()
		targetURL = fmt.Sprintf("%s:%s", imageURL, imageTag)
	)

	p := &pipeline{
		pid:       data.PID,
		service:   data.Service,
		appPath:   appPath,
		buildPath: buildPath,
		codePath:  codePath,
		imageURL:  imageURL,
		imageTag:  imageTag,
		targetURL: targetURL,
	}
	p.run(data)
}

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

func (p *pipeline) run(data Image) {
	util.Mkdir(p.buildPath) // 构建路径: 主路径/服务/上线单ID
	util.Mkdir(p.codePath)  // 代码路径: 主路径/服务/上线单ID/code
	log.Infof("current code path: %s", p.codePath)

	for _, item := range data.Build {
		if err := p.downloadCode(item.Module, item.Repo, item.Tag, p.codePath); err != nil {
			log.Errorf("download code failed: %+v", err)
			return
		}
	}

	if !p.compile(data.Type) {
		return
	}
	if !p.copyDockerfile() {
		return
	}
	if !p.dockerBuild() {
		return
	}
	if !p.dockerTag() {
		return
	}
	if !p.dockerPush() {
		return
	}
	if !p.UpdateImage() {
		return
	}
	log.Infof("push image: %s to registry success.", p.targetURL)
}

func (p *pipeline) downloadCode(module, repo, tag, codePath string) error {
	directory := filepath.Join(codePath, module)
	util.Rmdir(directory)

	if err := util.ExecuteDir(codePath, "git", "clone", repo); err != nil {
		return err
	}

	if err := util.ExecuteDir(directory, "git", "checkout", tag); err != nil {
		return err
	}
	return nil
}

func (p *pipeline) compile(language string) bool {
	switch language {
	case PYTHON:
	case GOLANG:
	}
	return true
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
	log.Infof("copy dockerfile: %s success.", dstFile)
	return true
}

func (p *pipeline) dockerBuild() bool {
	si, err := model.GetServiceInfo(p.service)
	if err != nil {
		log.Errorf("query service: %s failed: %s", p.service, err)
		return false
	}

	var (
		pull  = fmt.Sprintf("docker pull %s", si.ImageAddr)
		repo  = fmt.Sprintf("repo=%s", si.ImageAddr)
		param = fmt.Sprintf("docker build --build-arg %s -t %s %s", repo, p.targetURL, p.buildPath)
	)

	log.Info(pull)
	if err := util.Execute(pull); err != nil {
		log.Errorf("docker pull error: %+v", err)
		return false
	}

	log.Info(param)
	if err := util.Execute(param); err != nil {
		log.Errorf("docker build error: %+v", err)
		return false
	}
	return true
}

func (p *pipeline) dockerTag() bool {
	param := fmt.Sprintf("docker tag %s %s", p.targetURL, p.targetURL)
	log.Info(param)
	if err := util.Execute(param); err != nil {
		log.Errorf("docker tag error: %+v", err)
		return false
	}
	return true
}

func (p *pipeline) dockerPush() bool {
	param := fmt.Sprintf("docker push %s", p.targetURL)
	log.Info(param)
	if err := util.Execute(param); err != nil {
		log.Errorf("docker push error: %+v", err)
		return false
	}
	return true
}

func (p *pipeline) UpdateImage() bool {
	if err := model.UpdateImage(p.pid, p.imageURL, p.imageTag); err != nil {
		log.Errorf("write image info to db error: %s", err)
		return false
	}
	return true
}
