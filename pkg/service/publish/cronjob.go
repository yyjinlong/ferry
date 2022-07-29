// copyright @ 2022 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/config"
	"nautilus/pkg/k8s/exec"
	"nautilus/pkg/k8s/yaml"
	"nautilus/pkg/model"
	"nautilus/pkg/util"
)

func NewCronjob() *Cronjob {
	return &Cronjob{}
}

type Cronjob struct{}

func (c *Cronjob) Handle(namespace, service, command, schedule string) (string, error) {
	crontabID, err := model.CreateCrontab(namespace, service, command, schedule)
	if err != nil {
		return "", fmt.Errorf(config.CRON_WRITE_DB_ERROR, err)
	}

	name := util.GetCronjobName(service, crontabID)
	log.Infof("generate cronjob name: %s", name)

	svcObj, err := model.GetServiceInfo(service)
	if err != nil {
		return "", fmt.Errorf(config.DB_SERVICE_QUERY_ERROR, err)
	}

	serviceID := svcObj.ID
	pipeline, err := model.GetServicePipeline(serviceID)
	if err != nil {
		return "", fmt.Errorf(config.DB_PIPELINE_QUERY_BY_SERVICE_ID_ERROR, serviceID, err)
	}
	pid := pipeline.ID
	imageInfo, err := model.FindImageInfo(pid)
	if err != nil {
		return "", fmt.Errorf(config.DB_PIPELINE_UPDATE_ERROR, err)
	}

	if len(imageInfo) == 0 {
		return "", fmt.Errorf("get image info is empty")
	}
	log.Infof("create cronjob yaml get image info: %s", imageInfo)

	cronYaml := &yaml.CronjobYaml{
		Namespace:   namespace,
		Service:     service,
		ImageURL:    imageInfo["image_url"],
		ImageTag:    imageInfo["image_tag"],
		VolumeConf:  svcObj.Volume,
		ReserveTime: svcObj.ReserveTime,
		Name:        name,
		Schedule:    schedule,
		Command:     command,
	}
	tpl, err := cronYaml.Instance()
	if err != nil {
		return "", fmt.Errorf(config.CRON_BUILD_YAML_ERROR, err)
	}
	log.Infof("generate cronjob yaml(%s) success", name)
	fmt.Println(tpl)

	if err := c.execute(namespace, name, tpl); err != nil {
		return "", fmt.Errorf(config.CRON_K8S_EXEC_FAILED, err)
	}
	log.Infof("create namespace: %s cronjob: %s success", namespace, name)
	return name, nil
}

func (c *Cronjob) execute(namespace, name, tpl string) error {
	cron := exec.NewCronjob(namespace, name)
	if !cron.Exist() {
		return cron.Create(tpl)
	}
	return cron.Update(tpl)
}

func NewCronjobDelete() *CronjobDelete {
	return &CronjobDelete{Logid: util.UniqueID()}
}

type CronjobDelete struct {
	Logid string
}

func (c *CronjobDelete) Handle(namespace, service string, jobID int64) error {
	name := util.GetCronjobName(service, jobID)
	log.Infof("delete cronjob name: %s", name)

	cron := exec.NewCronjob(namespace, name)
	if !cron.Exist() {
		log.Infof("cronjob: %s is not exist", name)
		return nil
	}
	err := cron.Delete()
	if err != nil {
		log.Errorf("delete cronjob: %s failed: %s", name, err)
		return err
	}
	log.Infof("delete cronjob: %s success", name)
	return nil
}
