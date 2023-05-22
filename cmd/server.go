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
	log "github.com/sirupsen/logrus"

	"nautilus/pkg/config"
	"nautilus/pkg/model"
	"nautilus/pkg/router"
)

var (
	configFile  = flag.String("c", "../etc/dev.yaml", "yaml configuration file.")
	gracePeriod = 2 * time.Second
)

func main() {
	flag.Parse()
	config.ParseConfig(*configFile)
	config.InitLogger(config.Config().Log.Server)

	model.Connect("postgres",
		config.Config().Postgres.Master,
		config.Config().Postgres.Slave1,
		config.Config().Postgres.Slave2)

	r := gin.Default()
	r.Use(cors.Default())
	router.URLs(r)

	ctx, cancel := context.WithCancel(context.Background())

	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	server := http.Server{
		Addr:    config.Config().Address,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err == http.ErrServerClosed {
			log.Infof("Stopping everything, waiting %s...", gracePeriod)
		} else if err != nil {
			log.Errorf("listen and serve failed: %s", err)
			cancel()
		}
	}()

	select {
	case <-quitSignal:
		log.Infof("Signal(ctrl-c) captured, exiting...")
		server.Shutdown(ctx)
		cancel()
	}

	time.Sleep(gracePeriod)
}
