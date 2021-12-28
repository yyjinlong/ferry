// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"flag"

	"ferry/internal/bll/image"
	"ferry/internal/bll/listen"
	"ferry/pkg/db"
	"ferry/pkg/g"
	"ferry/pkg/log"
)

var (
	cfgFile = flag.String("c", "../../etc/dev.yaml", "yaml configuration file.")
	help    = flag.Bool("h", false, "show help info.")
)

func main() {
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}
	g.ParseConfig(*cfgFile)

	log.InitLogger(g.Config().Build.LogFile)

	db.Connect()

	go image.ListenMQ()
	go image.HandleMsg()

	go listen.DeploymentFinishEvent()
	go listen.EndpointFinishEvent()

	done := make(chan int)
	<-done
}
