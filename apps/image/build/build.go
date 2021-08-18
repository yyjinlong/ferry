// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package build

import (
	"ferry/ops/base"
	"ferry/ops/g"
)

type builder interface {
	getBuildPath(pid int64, mainPath, service string) string // 返回构建路径
	getCodePath(buildPath string) string                     // 返回代码路径
	download(codePath, module, repo, tag string) error       // 下载代码
	dockerBuild(pid int64, buildPath, service string) error  // docker build ...
}

func handler(br builder, data base.Image) {
	pid := data.PID
	service := data.Service

	buildPath := br.getBuildPath(pid, g.Config().Build.Dir, service)
	for _, item := range data.Build {
		codePath := br.getCodePath(buildPath)
		if err := br.download(codePath, item.Module, item.Repo, item.Tag); err != nil {
			return
		}
	}
	br.dockerBuild(pid, buildPath, service)
}
