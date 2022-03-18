// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"flag"

	"github.com/yyjinlong/golib/db"
	"github.com/yyjinlong/golib/log"

	"nautilus/pkg/config"
	"nautilus/pkg/service/listen"
)

var (
	configFile = flag.String("c", "../etc/dev.yaml", "yaml configuration file.")
	help       = flag.Bool("h", false, "show help info.")
)

func main() {
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}
	config.ParseConfig(*configFile)

	log.InitLogger(config.Config().Build.CronFile)

	db.Connect("postgres",
		config.Config().Postgres.Master,
		config.Config().Postgres.Slave1,
		config.Config().Postgres.Slave2)

	go listen.DeploymentFinishEvent()
	go listen.PublishLogEvent()

	done := make(chan int)
	<-done
}
