// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package image

const (
	PYTHON = "python"
	GOLANG = "go"
)

// 构建镜像
type Image struct {
	PID     int64        `json:"pid"`
	Type    string       `json:"type"`
	Service string       `json:"service"`
	Build   []ModuleInfo `json:"build"`
}

type ModuleInfo struct {
	Module string `json:"module"`
	Repo   string `json:"repo"`
	Tag    string `json:"tag"`
}
