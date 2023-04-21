// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"flag"

	"nautilus/cmd/informer/event"
	"nautilus/pkg/config"
	"nautilus/pkg/model"
)

var (
	configFile = flag.String("c", "../../etc/dev.yaml", "yaml configuration file.")
	cluster    = flag.String("s", "hp", "cluster param.")
	help       = flag.Bool("h", false, "show help info.")
)

func main() {
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}
	config.ParseConfig(*configFile)
	config.InitLogger(config.Config().Log.Informer)

	model.Connect("postgres",
		config.Config().Postgres.Master,
		config.Config().Postgres.Slave1,
		config.Config().Postgres.Slave2)

	clientset, err := config.GetClientset(*cluster)
	if err != nil {
		panic(err)
	}

	e := event.NewEvent(clientset)
	go event.DeploymentEvent(e, *cluster, clientset)
	go event.EndpointEvent(e, *cluster, clientset)
	go event.CronjobEvent(e, *cluster, clientset)
	go event.LogEvent(e, *cluster, clientset)

	done := make(chan int)
	<-done
}
