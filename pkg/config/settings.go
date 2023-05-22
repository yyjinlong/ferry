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

const (
	HP = "hp" // HP k8s cluster: hp
	XQ = "xq" // XQ k8s cluster: xq
)

type Settings struct {
	Address  string       `yaml:"address"`
	Log      LogInfo      `yaml:"log"`
	Postgres PostgresInfo `yaml:"postgres"`
	K8S      K8SInfo      `yaml:"k8s"`
}

type LogInfo struct {
	Server   string `yaml:"server"`
	Informer string `yaml:"informer"`
}

type PostgresInfo struct {
	Master string `yaml:"master"`
	Slave1 string `yaml:"slave1"`
	Slave2 string `yaml:"slave2"`
}

type K8SInfo struct {
	HP       string `yaml:"hp"`
	XQ       string `yaml:"xq"`
	ImageKey string `yaml:"imageKey"`
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

// InitLogger create and set logger
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

// GetClientset get client-go clientset
func GetClientset(cluster string) (*kubernetes.Clientset, error) {
	var clusterConfig string
	switch cluster {
	case HP:
		clusterConfig = setting.K8S.HP
	case XQ:
		clusterConfig = setting.K8S.XQ
	default:
		return nil, fmt.Errorf("unknown cluster: %s", cluster)
	}

	configFile, err := ioutil.ReadFile(clusterConfig)
	if err != nil {
		return nil, err
	}

	kubeConfig, err := clientcmd.RESTConfigFromKubeConfig(configFile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
