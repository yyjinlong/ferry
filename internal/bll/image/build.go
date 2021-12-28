// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package mirror

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"ferry/pkg/g"
	"ferry/pkg/log"
	"ferry/server/objects"
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
		g.DownloadCode(item.Module, item.Repo, item.Tag, p.codePath)
	}

	p.compile(data.Type)
	p.copyDockerfile()
	p.dockerBuild()
	p.dockerTag()
	p.dockerPush()
	p.writeImageToDB()
	log.Infof("push image: %s to registry success.", p.targetURL)
}

func (p *pipeline) compile(language string) {
	switch language {
	case PYTHON:
	case GOLANG:
	}
}

func (p *pipeline) copyDockerfile() {
	var (
		srcFile = filepath.Join(p.appPath, "mirror", "Dockerfile")
		dstFile = filepath.Join(p.buildPath, "Dockerfile")
	)
	if err := g.Copy(srcFile, dstFile); err != nil {
		log.Errorf("copy dockerfile: %s failed: %s", srcFile, err)
		return
	}
	log.Infof("copy dockerfile: %s success.", dstFile)
}

func (p *pipeline) dockerBuild() {
	si, err := objects.GetServiceInfo(p.service)
	if err != nil {
		log.Errorf("query service: %s failed: %s", p.service, err)
		return
	}

	var (
		repo = fmt.Sprintf("repo=%s", si.Service.ImageAddr)
		cmd  = fmt.Sprintf("docker build --build-arg %s -t %s %s", repo, p.targetURL, p.buildPath)
	)

	log.Info(cmd)
	if err := g.Execute("/bin/bash", "-c", cmd); err != nil {
		log.Errorf("docker build error: %s", err)
		return
	}
}

func (p *pipeline) dockerTag() {
	cmd := fmt.Sprintf("docker tag %s %s", p.targetURL, p.targetURL)
	log.Info(cmd)
	if err := g.Execute("/bin/bash", "-c", cmd); err != nil {
		log.Errorf("docker tag error: %s", err)
		return
	}
}

func (p *pipeline) dockerPush() {
	cmd := fmt.Sprintf("docker push %s", p.targetURL)
	log.Info(cmd)
	if err := g.Execute("/bin/bash", "-c", cmd); err != nil {
		log.Errorf("docker push error: %s", err)
		return
	}
}

func (p *pipeline) writeImageToDB() {
	if err := objects.CreateImage(p.pid, p.imageURL, p.imageTag); err != nil {
		log.Errorf("write image info to db error: %s", err)
		return
	}
}
