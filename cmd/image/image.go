// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"flag"

	"nautilus/cmd/image/app"
	"nautilus/pkg/config"
	"nautilus/pkg/model"
)

var (
	configFile = flag.String("c", "../../etc/dev.yaml", "yaml configuration file.")
	help       = flag.Bool("h", false, "show help info.")
)

func main() {
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}
	config.ParseConfig(*configFile)
	config.InitLogger(config.Config().Image.LogFile)

	model.Connect("postgres",
		config.Config().Postgres.Master,
		config.Config().Postgres.Slave1,
		config.Config().Postgres.Slave2)

	go app.ListenMQ()
	go app.HandleMsg()

	done := make(chan int)
	<-done
}
