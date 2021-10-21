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

func Execute(data Image) {
	p := &Pipeline{
		pid:      data.PID,
		service:  data.Service,
		imageURL: fmt.Sprintf("%s/%s", g.Config().Registry.Release, data.Service),
		imageTag: fmt.Sprintf("v-%s", time.Now().Format("20060102_150405")),
	}
	p.appPath = filepath.Dir(filepath.Dir(p.getCurPath()))
	p.buildPath = filepath.Join(g.Config().Build.Dir, p.service, strconv.FormatInt(p.pid, 10))
	p.codePath = filepath.Join(p.buildPath, "code")
	p.targetURL = fmt.Sprintf("%s:%s", p.imageURL, p.imageTag)
	p.Run(data)
}

type Pipeline struct {
	pid       int64
	service   string
	appPath   string
	buildPath string
	codePath  string
	imageURL  string
	imageTag  string
	targetURL string
}

func (p *Pipeline) getCurPath() string {
	_, curPath, _, _ := runtime.Caller(1)
	return curPath
}

func (p *Pipeline) Run(data Image) {
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
	if err := objects.CreateImage(p.pid, p.imageURL, p.imageTag); err != nil {
		log.Errorf("write image info to db error: %s", err)
		return
	}
	log.Info("release image: %s to registry success.", p.targetURL)
}

func (p *Pipeline) compile(language string) {
	switch language {
	case PYTHON:
	case GOLANG:
	}
}

func (p *Pipeline) copyDockerfile() {
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

func (p *Pipeline) dockerBuild() {
	si, err := objects.GetServiceInfo(p.service)
	if err != nil {
		log.Errorf("query service: %s failed: %s", p.service, err)
		return
	}

	var (
		repo = fmt.Sprintf("repo=%s", si.Service.ImageAddr)
		cmd  = fmt.Sprintf("docker build --build-arg %s -t %s %s", repo, p.targetURL, p.buildPath)
	)
	log.Infof(cmd)
	if err := g.Execute("/bin/bash", "-c", cmd); err != nil {
		log.Errorf("docker build error: %s", err)
		return
	}
}

func (p *Pipeline) dockerTag() {
	cmd := fmt.Sprintf("docker tag %s %s", p.targetURL, p.targetURL)
	log.Infof(cmd)
	if err := g.Execute("/bin/bash", "-c", cmd); err != nil {
		log.Errorf("docker tag error: %s", err)
		return
	}
}

func (p *Pipeline) dockerPush() {
	cmd := fmt.Sprintf("docker push %s", p.targetURL)
	log.Infof(cmd)
	if err := g.Execute("/bin/bash", "-c", cmd); err != nil {
		log.Errorf("docker push error: %s", err)
		return
	}
}
