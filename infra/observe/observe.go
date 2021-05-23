package pkg

import (
	"fmt"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/ocagent"
	conf "github.com/minghsu0107/saga-account/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

type ObservibilityInjector struct {
	promPort    string
	ocagentHost string
	appName     string
}

func NewObservibilityInjector(config *conf.Config) (*ObservibilityInjector, error) {
	promPort := config.PromPort
	ocagentHost := config.OcAgentHost
	appName := config.AppName

	if appName == "" {
		return nil, fmt.Errorf("app name should not be empty")
	}

	return &ObservibilityInjector{
		promPort:    promPort,
		ocagentHost: ocagentHost,
		appName:     appName,
	}, nil
}

func (injector *ObservibilityInjector) Register(errs chan error) {
	if injector.ocagentHost != "" {
		oce, err := ocagent.NewExporter(
			ocagent.WithInsecure(),
			ocagent.WithReconnectionPeriod(5*time.Second),
			ocagent.WithAddress(injector.ocagentHost),
			ocagent.WithServiceName(injector.appName))
		if err != nil {
			log.Fatalf("failed to create ocagent-exporter: %v", err)
		}
		trace.RegisterExporter(oce)
		/*
			// if parant span is sampled, the current is also sampled
			// despite the sampling configuration in order to obtain full span tree
			trace.ApplyConfig(trace.Config{
				// If not specified, then sampler would be set to ProbabilitySampler(defaultSamplingProbability)
				// defaultSamplingProbability is 1e-4
				// DefaultSampler: trace.NeverSample(),
			})
		*/
	}
	if injector.promPort != "" {
		go func() {
			log.Infof("starting prom metrics on PROM_PORT=[%s]", injector.promPort)
			errs <- http.ListenAndServe(fmt.Sprintf(":%s", injector.promPort), promhttp.Handler())
		}()
	}
}
