// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package build

import (
	"ferry/imager/model"
	"ferry/ops/g"
)

/*
 * GetBuildPath 返回镜像构建路径(主路径/服务/服务ID) 如: /tmp/release/ivr/8
 * GetCodePath  返回镜像代码路径(主路径/服务/服务ID/codela. 如: ) 如: /tmp/release/ivr/8/code
 * DownloadCode 下载对应tag的代码 如: /tmp/release/ivr/8/code/ivr
 * ReleaseImage 构建镜像、镜像打tag、镜像push
 *
 * 目录如下:
 * ➜  ~ ls /tmp/release/ivr/8/
 * Dockerfile code
 *
 */
type Builder interface {
	GetBuildPath(mainPath, service string, pid int64) string // 返回构建路径
	GetCodePath(buildPath string) string                     // 返回代码路径
	DownloadCode(codePath, module, repo, tag string) error   // 下载代码
	ReleaseImage(buildPath, service string, pid int64) error // 构建及发布镜像
}

func handler(br Builder, data model.Image) {
	var (
		pid     = data.PID
		service = data.Service
	)

	buildPath := br.GetBuildPath(g.Config().Build.Dir, service, pid)
	for _, item := range data.Build {
		codePath := br.GetCodePath(buildPath)
		if err := br.DownloadCode(codePath, item.Module, item.Repo, item.Tag); err != nil {
			return
		}
	}
	br.ReleaseImage(buildPath, service, pid)
}
