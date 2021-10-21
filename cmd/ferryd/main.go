// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"flag"

	"ferry/cmd/ferryd/mirror"
	"ferry/cmd/ferryd/trace"
	"ferry/pkg/g"
	"ferry/pkg/log"
	"ferry/server/db"
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

	go mirror.ListenMQ()
	go mirror.HandlePy()
	go mirror.HandleGo()

	ee := trace.NewEvent()
	go ee.Deployment()
	go ee.Endpoint()

	done := make(chan int)
	<-done
}
