package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/minghsu0107/saga-account/dep"
	"github.com/minghsu0107/saga-account/infra/cache"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

var (
	promPort    = os.Getenv("PROM_PORT")
	ocagentHost = os.Getenv("OC_AGENT_HOST")
)

func main() {
	errs := make(chan error, 1)
	if ocagentHost != "" {
		oce, err := ocagent.NewExporter(
			ocagent.WithInsecure(),
			ocagent.WithReconnectionPeriod(5*time.Second),
			ocagent.WithAddress(ocagentHost),
			ocagent.WithServiceName("account"))
		if err != nil {
			log.Fatalf("failed to create ocagent-exporter: %v", err)
		}
		trace.RegisterExporter(oce)
		trace.ApplyConfig(trace.Config{
			// If not specify then sampler would be set to ProbabilitySampler(defaultSamplingProbability)
			//DefaultSampler: trace.ProbabilitySampler(0.2),
			DefaultSampler: trace.AlwaysSample(),
		})
	}
	if promPort != "" {
		go func() {
			log.Infof("starting prom metrics on PROM_PORT=[%s]", promPort)
			http.Handle("/metrics", promhttp.Handler())
			err := http.ListenAndServe(fmt.Sprintf(":%s", promPort), nil)
			errs <- err
		}()
	}

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
