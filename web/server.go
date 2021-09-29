// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

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

	log.InitLogger(g.Config().LogFile)

	db.Connect()

	ctx, cancel := context.WithCancel(context.Background())

	qs := make(chan os.Signal)
	signal.Notify(qs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	r := gin.Default()
	urls(r)

	server := http.Server{
		Addr:           g.Config().Address,
		Handler:        r,
		ReadTimeout:    time.Duration(g.Config().ReadTimeout),
		WriteTimeout:   time.Duration(g.Config().WriteTimeout),
		MaxHeaderBytes: g.Config().MaxHeaderBytes,
	}

	go func() {
		if err := server.ListenAndServe(); err == http.ErrServerClosed {
			log.Info("listen and serve shutdown....")
		} else if err != nil {
			log.Errorf("listen and serve failed: %s", err)
			cancel()
		}
	}()

	select {
	case sig := <-qs:
		if sig == syscall.SIGINT || sig == syscall.SIGTERM || sig == syscall.SIGQUIT {
			log.Info("quit the server with ctrl c")
			server.Shutdown(ctx)
			cancel()
		}
	}

	time.Sleep(time.Duration(g.Config().ExitWaitSecond) * time.Second)
}
