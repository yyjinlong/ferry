// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package exec

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/config"
	"nautilus/pkg/util/curl"
)

func NewCronjob(namespace, name string) *Cronjob {
	cluster := getCluster(namespace)
	return &Cronjob{
		cluster:   cluster,
		address:   getAddress(cluster),
		namespace: namespace,
		name:      name,
	}
}

type Cronjob struct {
	cluster   string
	address   string
	namespace string
	name      string
}

func (cj *Cronjob) Exist() bool {
	var (
		url    = fmt.Sprintf(config.Config().K8S.Cronjob, cj.address, cj.namespace) + "/" + cj.name
		header = getHeader(cj.cluster)
	)

	body, err := curl.Get(url, header, 5)
	if err != nil {
		log.Infof("check cronjob: %s is not exist", cj.name)
		return false
	}
	if err := response(body); err != nil {
		return false
	}
	log.Infof("check cronjob: %s is exist", cj.name)
	return true
}

func (cj *Cronjob) Create(tpl string) error {
	var (
		url    = fmt.Sprintf(config.Config().K8S.Cronjob, cj.address, cj.namespace)
		header = getHeader(cj.cluster)
	)
	body, err := curl.Post(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request create cronjob api error: %s", err)
		return err
	}
	return response(body)
}

func (cj *Cronjob) Update(tpl string) error {
	var (
		url    = fmt.Sprintf(config.Config().K8S.Cronjob, cj.address, cj.namespace) + "/" + cj.name
		header = getHeader(cj.cluster)
	)
	body, err := curl.Put(url, header, []byte(tpl), 5)
	if err != nil {
		log.Errorf("request update cronjob api error: %s", err)
		return err
	}
	return response(body)
}

func (cj *Cronjob) Delete() error {
	var (
		url    = fmt.Sprintf(config.Config().K8S.Cronjob, cj.address, cj.namespace) + "/" + cj.name
		header = getHeader(cj.cluster)
	)
	body, err := curl.Delete(url, header, 5)
	if err != nil {
		log.Errorf("request delete cronjob api error: %s", err)
		return err
	}
	return response(body)
}
