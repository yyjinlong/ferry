// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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
	HPConfig   string `yaml:"hpconfig"`
	XQConfig   string `yaml:"xqconfig"`
	Deployment string `yaml:"deployment"`
	Service    string `yaml:"service"`
	ConfigMap  string `yaml:"configmap"`
	Cronjob    string `yaml:"cronjob"`
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

// -----------log配置--------------
func InitLogger(logFile string) {
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	// 同时写文件和屏幕
	writers := []io.Writer{f, os.Stdout}
	allWriters := io.MultiWriter(writers...)

	log.SetReportCaller(true)
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			filename := path.Base(frame.File)
			return frame.Function, filename + ":" + strconv.Itoa(frame.Line)
		},
	})
	log.SetOutput(allWriters)
	log.SetLevel(log.InfoLevel)
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

func GetClientset(cluster string) (*kubernetes.Clientset, error) {
	var clusterConfig string
	if cluster == HP {
		clusterConfig = setting.K8S.HPConfig
	} else if cluster == XQ {
		clusterConfig = setting.K8S.XQConfig
	} else {
		return nil, fmt.Errorf("unknown cluster: %s", cluster)
	}

	configFile, err := ioutil.ReadFile(clusterConfig)
	if err != nil {
		return nil, err
	}

	kubeconfig, err := clientcmd.RESTConfigFromKubeConfig(configFile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
