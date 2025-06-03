package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/DazWilkin/go-probe/probe"

	"github.com/DazWilkin/go-usgs-waterdata/waterdata"

	"github.com/DazWilkin/go-usgs-waterdata-exporter/collector"
)

const (
	namespace string = "usgs_waterdata"
	subsystem string = "exporter"
	version   string = "v0.0.1"
)

var (
	// GitCommit is the git commit value and is expected to be set during build
	GitCommit string
	// GoVersion is the Golang runtime version
	GoVersion = runtime.Version()
	// OSVersion is the OS version (uname --kernel-release) and is expected to be set during build
	OSVersion string
	// StartTime is the start time of the exporter represented as a UNIX epoch
	StartTime = time.Now().Unix()
)
var (
	endpoint    = flag.String("endpoint", "0.0.0.0:8080", "The endpoint of the Prometheus exporter")
	metricsPath = flag.String("path", "/metrics", "The path on which Prometheus metrics will be served")
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	var sitecodes []string
	flag.Func("sitecode", "The sitecodes to be queried", func(s string) error {
		sitecodes = append(sitecodes, s)
		return nil
	})

	flag.Parse()

	if *endpoint == "" {
		logger.Error("expected flag `--endpoint`")
		os.Exit(1)
	}
	if len(sitecodes) == 0 {
		logger.Error("expected at least one flag `--sitecode`")
		os.Exit(1)
	}

	if GitCommit == "" {
		logger.Info("GitCommit value unchanged: expected to be set during build")
	}
	if OSVersion == "" {
		logger.Info("OSVersion value unchanged: expected to be set during build")
	}

	client, err := waterdata.NewClient(logger)
	if err != nil {
		logger.Error("unable to create client",
			"err", err,
		)
		os.Exit(1)
	}

	p := probe.New("liveness", logger)
	healthz := p.Handler(logger)

	// Channel is shared by the Updater (subscriber) and the InstananeousValuesCollector (publisher)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan probe.Status)
	go p.Updater(ctx, ch, nil)

	s := collector.System{
		Namespace: namespace,
		Subsystem: subsystem,
		Version:   version,
	}

	b := collector.Build{
		OsVersion: OSVersion,
		GoVersion: GoVersion,
		GitCommit: GitCommit,
		StartTime: StartTime,
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewExporterCollector(s, b, logger))
	registry.MustRegister(collector.NewInstantaneousValuesCollector(s, client, ch, sitecodes, logger))

	mux := http.NewServeMux()
	mux.Handle("/", root(logger))
	mux.Handle("/healthz", healthz)
	mux.Handle("/robots.txt", robots(logger))

	mux.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	logger.Info("Starting",
		"endpoint", *endpoint,
		"metrics", *metricsPath,
	)
	logger.Error("unable to start server",
		"err", http.ListenAndServe(*endpoint, mux),
	)
}
