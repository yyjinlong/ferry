// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"flag"

	"nautilus/cmd/informer/app"
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
	config.InitLogger(config.Config().Informer.LogFile)

	model.Connect("postgres",
		config.Config().Postgres.Master,
		config.Config().Postgres.Slave1,
		config.Config().Postgres.Slave2)

	clientset := app.GetClientset(*cluster)
	go app.DeploymentFinishEvent(clientset)
	go app.PublishLogEvent(clientset)
	go app.EndpointFinishEvent(clientset)
	go app.CronjobFinishEvent(clientset)

	done := make(chan int)
	<-done
}
