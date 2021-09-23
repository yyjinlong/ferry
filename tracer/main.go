// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"flag"
	"io/ioutil"

	"ferry/ops/db"
	"ferry/ops/g"
	"ferry/ops/log"
	"ferry/tracer/listen"
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

	log.InitLogger(g.Config().Watch.LogFile)

	db.Connect()

	kubeconfig, err := ioutil.ReadFile(g.Config().K8S.Kubeconfig)
	if err != nil {
		log.Panicf("read kubeconfig file error: %s", err)
	}
	event := listen.NewEvent(kubeconfig)
	go event.Deployment()
	go event.Endpoint()

	done := make(chan int)
	<-done
}
