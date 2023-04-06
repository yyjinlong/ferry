package k8s

import (
	"fmt"
)

const (
	BLUE  = "blue"
	GREEN = "green"
)

// GetDeploymentName 生成deployment名字 规则: 服务名-服务ID-部署阶段-部署组
func GetDeploymentName(serviceName string, serviceID int64, phase, group string) string {
	return fmt.Sprintf("%s-%d-%s-%s", serviceName, serviceID, phase, group)
}

// GetAppID 生成appid 规则: 服务名-服务ID-部署阶段
func GetAppID(serviceName string, serviceID int64, phase string) string {
	return fmt.Sprintf("%s-%d-%s", serviceName, serviceID, phase)
}

// GetDeployGroup 获取部署组
func GetDeployGroup(onlineGroup string) string {
	if onlineGroup == BLUE {
		return GREEN
	}
	return BLUE
}

// GetAnotherGroup 获取当前组对应的另一组
func GetAnotherGroup(group string) string {
	return GetDeployGroup(group)
}

// GetConfigmapName 生成configmap名字 规则: 服务名-config
func GetConfigmapName(serviceName string) string {
	return fmt.Sprintf("%s-config", serviceName)
}

// GetCronjobName 生成cronjob名字 规则: 服务名-cronjob-任务ID
func GetCronjobName(serviceName string, jobID int64) string {
	return fmt.Sprintf("%s-cronjob-%d", serviceName, jobID)
}
