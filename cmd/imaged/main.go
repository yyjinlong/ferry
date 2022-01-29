// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"flag"

	"nautilus/internal/bll/image"
	"nautilus/pkg/db"
	"nautilus/pkg/g"
	"nautilus/pkg/log"
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

	log.InitLogger(g.Config().Build.ImgFile)

	db.Connect()

	go image.ListenMQ()
	go image.HandleMsg()

	done := make(chan int)
	<-done
}
