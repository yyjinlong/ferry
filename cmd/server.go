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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/yyjinlong/golib/db"
	"github.com/yyjinlong/golib/log"

	"nautilus/pkg/config"
	"nautilus/pkg/url"
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

	log.InitLogger(config.Config().LogFile)

	db.Connect("postgres",
		config.Config().Postgres.Master,
		config.Config().Postgres.Slave1,
		config.Config().Postgres.Slave2)

	ctx, cancel := context.WithCancel(context.Background())

	qs := make(chan os.Signal)
	signal.Notify(qs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	r := gin.Default()
	r.Use(cors.Default())
	url.URLs(r)

	server := http.Server{
		Addr:           config.Config().Address,
		Handler:        r,
		ReadTimeout:    time.Duration(config.Config().ReadTimeout),
		WriteTimeout:   time.Duration(config.Config().WriteTimeout),
		MaxHeaderBytes: config.Config().MaxHeaderBytes,
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

	time.Sleep(time.Duration(config.Config().ExitWaitSecond) * time.Second)
}
