package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

var (
	promPort    = os.Getenv("PROM_PORT")
	ocagentHost = os.Getenv("OC_AGENT_HOST")
)

func main() {
	oce, err := ocagent.NewExporter(
		ocagent.WithInsecure(),
		ocagent.WithReconnectionPeriod(5*time.Second),
		ocagent.WithAddress(ocagentHost),
		ocagent.WithServiceName("account-svc"))
	if err != nil {
		log.Fatalf("failed to create ocagent-exporter: %v", err)
	}
	trace.RegisterExporter(oce)

	errs := make(chan error, 1)
	if promPort != "" {
		// Start prometheus server
		go func() {
			log.Infof("starting prom metrics on PROM_PORT=[%s]", promPort)
			http.Handle("/metrics", promhttp.Handler())
			err := http.ListenAndServe(fmt.Sprintf(":%s", promPort), nil)
			errs <- err
		}()
	}
}
