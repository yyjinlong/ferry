// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"flag"

	"ferry/imager/build"
	"ferry/ops/db"
	"ferry/ops/g"
	"ferry/ops/log"
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
	g.ParseConfig(*cfgFile)

	log.InitLogger(g.Config().Build.LogFile)

	db.Connect()

	build.ListenImage()
}
