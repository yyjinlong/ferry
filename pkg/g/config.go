// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package g

import (
	"io/ioutil"
	"sync"

	yaml "gopkg.in/yaml.v2"
)

type Settings struct {
	Address        string       `yaml:"address"`
	ReadTimeout    int          `yaml:"readTimeout"`
	WriteTimeout   int          `yaml:"writeTimeout"`
	MaxHeaderBytes int          `yaml:"maxHeaderBytes"`
	ExitWaitSecond int          `yaml:"exitWaitSecond"`
	LogFile        string       `yaml:"logfile"`
	Postgres       PostgresInfo `yaml:"postgres"`
	RabbitMQ       RabbtimqInfo `yaml:"rabbitmq"`
	Build          BuildInfo    `yaml:"build"`
	Registry       RegistryInfo `yaml:"registry"`
	K8S            K8SInfo      `yaml:"k8s"`
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

type BuildInfo struct {
	Dir      string `yaml:"dir"`
	ImgFile  string `yaml:"imgfile"`
	CronFile string `yaml:"cronfile"`
}

type RegistryInfo struct {
	Release string `yaml:"release"`
}

type K8SInfo struct {
	Kubeconfig string `yaml:"kubeconfig"`
	Deployment string `yaml:"deployment"`
	Service    string `yaml:"service"`
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
