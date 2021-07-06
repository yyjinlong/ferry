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
	LogFile   string        `yaml:"logfile"`
	Bootstrap BootstrapInfo `yaml:"bootstrap"`
	Postgres  PostgresInfo  `yaml:"postgres"`
	K8S       K8SInfo       `yaml:"k8s"`
}

type BootstrapInfo struct {
	Address        string `yaml:"address"`
	ReadTimeout    int    `yaml:"readTimeout"`
	WriteTimeout   int    `yaml:"writeTimeout"`
	MaxHeaderBytes int    `yaml:"maxHeaderBytes"`
	ExitWaitSecond int    `yaml:"exitWaitSecond"`
}

type PostgresInfo struct {
	Master string `yaml:"master"`
	Slave1 string `yaml:"slave1"`
	Slave2 string `yaml:"slave2"`
}

type K8SInfo struct {
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
