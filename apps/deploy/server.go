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

	log.InitLogger(g.Config().LogFile)

	db.Connect()

	ctx, cancel := context.WithCancel(context.Background())

	qs := make(chan os.Signal)
	signal.Notify(qs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	r := gin.Default()
	urls(r)

	server := http.Server{
		Addr:           g.Config().Bootstrap.Address,
		Handler:        r,
		ReadTimeout:    time.Duration(g.Config().Bootstrap.ReadTimeout),
		WriteTimeout:   time.Duration(g.Config().Bootstrap.WriteTimeout),
		MaxHeaderBytes: g.Config().Bootstrap.MaxHeaderBytes,
	}

	go func() {
		if err := server.ListenAndServe(); err == http.ErrServerClosed {
			log.Info("Listen and serve shutdown....")
		} else if err != nil {
			log.Errorf("Listen and serve boot failed: %s", err)
			cancel()
		}
	}()

	select {
	case sig := <-qs:
		if sig == syscall.SIGINT || sig == syscall.SIGTERM || sig == syscall.SIGQUIT {
			log.Info("Quit the server with Ctrl C.")
			server.Shutdown(ctx)
			cancel()
		} else if sig == syscall.SIGPIPE {
			log.Warn("Ignore broken pipe signal.")
		}
	}

	time.Sleep(time.Duration(g.Config().Bootstrap.ExitWaitSecond) * time.Second)
}
