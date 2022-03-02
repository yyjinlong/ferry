// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"flag"

	"github.com/yyjinlong/golib/db"
	"github.com/yyjinlong/golib/log"

	"nautilus/pkg/bll/image"
	"nautilus/pkg/cfg"
)

var (
	cfgFile = flag.String("c", "../etc/dev.yaml", "yaml configuration file.")
	help    = flag.Bool("h", false, "show help info.")
)

func main() {
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}
	cfg.ParseConfig(*cfgFile)

	log.InitLogger(cfg.Config().Build.ImgFile)

	db.Connect("postgres",
		cfg.Config().Postgres.Master,
		cfg.Config().Postgres.Slave1,
		cfg.Config().Postgres.Slave2)

	go image.ListenMQ()
	go image.HandleMsg()

	done := make(chan int)
	<-done
}
