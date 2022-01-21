// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package image

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"nautilus/internal/objects"
	"nautilus/pkg/g"
	"nautilus/pkg/git"
	"nautilus/pkg/log"
)

func getCurPath() string {
	_, curPath, _, _ := runtime.Caller(1)
	return curPath
}

func worker(data Image) {
	var (
		pid       = data.PID
		service   = data.Service
		buildPath = filepath.Join(g.Config().Build.Dir, service, strconv.FormatInt(pid, 10))
		appPath   = filepath.Dir(filepath.Dir(getCurPath()))
		codePath  = filepath.Join(buildPath, "code")
		imageURL  = fmt.Sprintf("%s/%s", g.Config().Registry.Release, service)
		imageTag  = fmt.Sprintf("v-%s", time.Now().Format("20060102_150405"))
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
	g.Mkdir(p.buildPath) // 构建路径: 主路径/服务/上线单ID
	g.Mkdir(p.codePath)  // 代码路径: 主路径/服务/上线单ID/code
	log.Infof("current code path: %s", p.codePath)

	for _, item := range data.Build {
		git.DownloadCode(item.Module, item.Repo, item.Tag, p.codePath)
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
	if !p.writeImageToDB() {
		return
	}
	log.Infof("push image: %s to registry success.", p.targetURL)
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
		srcFile = filepath.Join(p.appPath, "image", "Dockerfile")
		dstFile = filepath.Join(p.buildPath, "Dockerfile")
	)
	if err := g.Copy(srcFile, dstFile); err != nil {
		log.Errorf("copy dockerfile: %s failed: %s", srcFile, err)
		return false
	}
	log.Infof("copy dockerfile: %s success.", dstFile)
	return true
}

func (p *pipeline) dockerBuild() bool {
	si, err := objects.GetServiceInfo(p.service)
	if err != nil {
		log.Errorf("query service: %s failed: %s", p.service, err)
		return false
	}

	var (
		repo = fmt.Sprintf("repo=%s", si.Service.ImageAddr)
		cmd  = fmt.Sprintf("docker build --build-arg %s -t %s %s", repo, p.targetURL, p.buildPath)
	)

	log.Info(cmd)
	if err := g.Execute("/bin/bash", "-c", cmd); err != nil {
		log.Errorf("docker build error: %s", err)
		return false
	}
	return true
}

func (p *pipeline) dockerTag() bool {
	cmd := fmt.Sprintf("docker tag %s %s", p.targetURL, p.targetURL)
	log.Info(cmd)
	if err := g.Execute("/bin/bash", "-c", cmd); err != nil {
		log.Errorf("docker tag error: %s", err)
		return false
	}
	return true
}

func (p *pipeline) dockerPush() bool {
	cmd := fmt.Sprintf("docker push %s", p.targetURL)
	log.Info(cmd)
	if err := g.Execute("/bin/bash", "-c", cmd); err != nil {
		log.Errorf("docker push error: %s", err)
		return false
	}
	return true
}

func (p *pipeline) writeImageToDB() bool {
	if err := objects.CreateImage(p.pid, p.imageURL, p.imageTag); err != nil {
		log.Errorf("write image info to db error: %s", err)
		return false
	}
	return true
}
