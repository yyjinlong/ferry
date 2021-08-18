// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package build

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"ferry/ops/base"
	"ferry/ops/g"
	"ferry/ops/log"
	"ferry/ops/objects"
)

func Python(data base.Image) {
	py := &pyBuild{}
	handler(py, data)
}

type pyBuild struct {
}

func (p *pyBuild) getBuildPath(pid int64, mainPath, service string) string {
	buildPath := filepath.Join(mainPath, service, strconv.FormatInt(pid, 10))
	if _, err := os.Stat(buildPath); os.IsNotExist(err) {
		os.MkdirAll(buildPath, os.ModePerm)
	}
	log.Infof("build image directory: %s", buildPath)
	return buildPath
}

func (p *pyBuild) getCodePath(buildPath string) string {
	codePath := filepath.Join(buildPath, "code")
	if _, err := os.Stat(codePath); os.IsNotExist(err) {
		os.Mkdir(codePath, os.ModePerm)
	}
	log.Infof("download code directory: %s", codePath)
	return codePath
}

func (p *pyBuild) download(codePath, module, repo, tag string) error {
	git := g.NewGit(module, repo, tag, codePath)
	if err := git.Clone(); err != nil {
		log.Errorf("git clone code failed: %s", err)
		return err
	}
	log.Infof("git clone module: %s success", module)

	if err := git.CheckoutTag(); err != nil {
		log.Errorf("git checkout tag failed: %s", err)
		return err
	}
	log.Infof("git checkout tag: %s success", tag)
	return nil
}

func (p *pyBuild) dockerBuild(pid int64, buildPath, service string) error {
	// NOTE: 拷贝Dockerfile
	_, curPath, _, _ := runtime.Caller(1)
	appPath := filepath.Dir(filepath.Dir(curPath))
	srcFile := filepath.Join(appPath, "dockerfile", "Dockerfile")
	dstFile := filepath.Join(buildPath, "Dockerfile")
	if err := g.Copy(srcFile, dstFile); err != nil {
		log.Errorf("copy dockerfile: %s failed: %s", srcFile, err)
		return err
	}
	log.Infof("(1) copy dockerfile: %s success.", dstFile)

	// NOTE: 构建镜像
	svc, err := objects.GetService(service)
	if err != nil {
		log.Errorf("query service: %s failed: %s", service, err)
		return err
	}
	repo := fmt.Sprintf("repo=%s", svc.ImageAddr)
	imageURL := fmt.Sprintf("%s/%s", g.Config().Registry.Release, service)
	imageTag := fmt.Sprintf("v-%s", time.Now().Format("20060102_150405"))
	releaseTag := imageURL + ":" + imageTag
	log.Infof("docker build --build-arg %s -t %s %s", repo, releaseTag, buildPath)

	if _, err := g.Cmd("docker", "build", "--build-arg", repo, "-t", releaseTag, buildPath); err != nil {
		log.Errorf("docker build error: %s", err)
		return err
	}
	log.Info("(2) docker build success")

	// NOTE: 打tag
	log.Infof("docker tag %s %s", releaseTag, releaseTag)
	if _, err := g.Cmd("docker", "tag", releaseTag, releaseTag); err != nil {
		log.Errorf("docker tag error: %s", err)
		return err
	}
	log.Info("(3) docker tag success")

	// NOTE: 推送镜像
	log.Infof("docker push %s", releaseTag)
	if _, err := g.Cmd("docker", "push", releaseTag); err != nil {
		log.Errorf("docker push error: %s", err)
		return err
	}
	log.Info("(4) docker push success")

	// NOTE: 保存镜像信息
	if err := objects.CreateImage(pid, imageURL, imageTag); err != nil {
		return err
	}
	log.Info("(5) write image info to db success")
	return nil
}
