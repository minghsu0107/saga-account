package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minghsu0107/saga-account/dep"
	"github.com/minghsu0107/saga-account/infra/cache"
	log "github.com/sirupsen/logrus"
)

func main() {
	errs := make(chan error, 1)

	migrator, err := dep.InitializeMigrator()
	if err != nil {
		log.Fatal(err)
	}
	if err := migrator.Migrate(); err != nil {
		log.Fatal(err)
	}

	server, err := dep.InitializeServer()
	if err != nil {
		log.Fatal(err)
	}
	defer cache.RedisClient.Close()

	go func() {
		errs <- server.Run()
	}()

	// catch shutdown
	done := make(chan bool, 1)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		// graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.GracefulStop(ctx, done)
	}()

	err = <-errs
	if err != nil {
		log.Fatal(err)
	}

	// wait for graceful shutdown
	<-done
}
