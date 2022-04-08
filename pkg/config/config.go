// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package config

import (
	"io/ioutil"
	"sync"

	yaml "gopkg.in/yaml.v2"
)

type Settings struct {
	Address  string       `yaml:"address"`
	LogFile  string       `yaml:"logfile"`
	Postgres PostgresInfo `yaml:"postgres"`
	RabbitMQ RabbtimqInfo `yaml:"rabbitmq"`
	Image    ImageInfo    `yaml:"image"`
	Informer InformerInfo `yaml:"informer"`
	K8S      K8SInfo      `yaml:"k8s"`
	Cluster  ClusterInfo  `yaml:"cluster"`
}

type PostgresInfo struct {
	Master string `yaml:"master"`
	Slave1 string `yaml:"slave1"`
	Slave2 string `yaml:"slave2"`
}

type RabbtimqInfo struct {
	Address    string `yaml:"address"`
	Exchange   string `yaml:"exchange"`
	Queue      string `yaml:"queue"`
	RoutingKey string `yaml:"routingKey"`
}

type ImageInfo struct {
	LogFile  string `yaml:"logfile"`
	Registry string `yaml:"registry"`
	Dir      string `yaml:"dir"`
}

type InformerInfo struct {
	LogFile string `yaml:"logfile"`
}

type K8SInfo struct {
	Kubeconfig string `yaml:"kubeconfig"`
	Deployment string `yaml:"deployment"`
	Service    string `yaml:"service"`
	ConfigMap  string `yaml:"configmap"`
}

type ClusterInfo struct {
	HP string `yaml:"hp"`
	XQ string `yaml:"xq"`
}

var (
	setting Settings
	lock    = new(sync.RWMutex)
)

func ParseConfig(cfgFile string) {
	buf, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		panic(err)
	}

	lock.Lock()
	defer lock.Unlock()

	if err := yaml.Unmarshal(buf, &setting); err != nil {
		panic(err)
	}
}

func Config() Settings {
	lock.RLock()
	defer lock.RUnlock()
	return setting
}

// -----------多集群映射-----------
const (
	HP = "hp"
	XQ = "xq"
)

func GetAddress(idc string) string {
	mapping := map[string]string{
		HP: Config().Cluster.HP,
		XQ: Config().Cluster.XQ,
	}
	return mapping[idc]
}
