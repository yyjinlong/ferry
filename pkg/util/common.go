// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package util

import (
	"fmt"
	"io"
	"os"
	"time"
)

const (
	BLUE  = "blue"
	GREEN = "green"
)

// GetDeployment 根据服务名、服务ID、部署阶段、部署组 来命名deployment name
func GetDeployment(serviceName string, serviceID int64, phase, group string) string {
	return fmt.Sprintf("%s-%d-%s-%s", serviceName, serviceID, phase, group)
}

func GetAppID(serviceName string, serviceID int64, phase string) string {
	return fmt.Sprintf("%s-%d-%s", serviceName, serviceID, phase)
}

func GetDeployGroup(onlineGroup string) string {
	if onlineGroup == BLUE {
		return GREEN
	}
	return BLUE
}

func GetConfigName(serviceName string) string {
	return fmt.Sprintf("%s-config", serviceName)
}

func In(data string, dataList []string) bool {
	for _, item := range dataList {
		if data == item {
			return true
		}
	}
	return false
}

func Ini(num int, numList []int) bool {
	for _, n := range numList {
		if num == n {
			return true
		}
	}
	return false
}

func TimeString(curTime time.Time) string {
	return curTime.Format("2006-01-02 15:04:05")
}

func Mkdir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}
}

func Copy(source, dest string) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
}
