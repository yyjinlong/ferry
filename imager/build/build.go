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
	getBuildPath(mainPath, service string, pid int64) string // 返回构建路径
	getCodePath(buildPath string) string                     // 返回代码路径
	download(codePath, module, repo, tag string) error       // 下载代码
	release(buildPath, service string, pid int64) error      // 构建及发布镜像
}

func handler(br builder, data base.Image) {
	pid := data.PID
	service := data.Service

	buildPath := br.getBuildPath(g.Config().Build.Dir, service, pid)
	for _, item := range data.Build {
		codePath := br.getCodePath(buildPath)
		if err := br.download(codePath, item.Module, item.Repo, item.Tag); err != nil {
			return
		}
	}
	br.release(buildPath, service, pid)
}
