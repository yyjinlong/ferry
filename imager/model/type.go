// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package model

// 构建镜像
type Image struct {
	PID     int64        `json:"pid"`
	Type    string       `json:"type"`
	Service string       `json:"service"`
	Build   []moduleInfo `json:"build"`
}

type moduleInfo struct {
	Module string `json:"module"`
	Repo   string `json:"repo"`
	Tag    string `json:"tag"`
}

const (
	PYTHON = "python"
	GOLANG = "go"
)
