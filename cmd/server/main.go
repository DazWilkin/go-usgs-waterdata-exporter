package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/DazWilkin/go-usgs-waterdata/collector"
	"github.com/DazWilkin/go-usgs-waterdata/waterdata"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	var sites []string
	flag.Func("site", "The sites to be queried", func(s string) error {
		sites = append(sites, s)
		return nil
	})

	flag.Parse()

	if *endpoint == "" {
		logger.Error("Expected flag `--endpoint`")
		os.Exit(1)
	}
	if len(sites)==0 {
		logger.Error("Expected at least one flag `--site`")
		os.Exit(1)
	}

	if GitCommit == "" {
		logger.Info("GitCommit value unchanged: expected to be set during build")
	}
	if OSVersion == "" {
		logger.Info("OSVersion value unchanged: expected to be set during build")
	}

	client := waterdata.NewClient(logger)

	registry := prometheus.NewRegistry()

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
	registry.MustRegister(collector.NewExporterCollector(s, b, logger))
	registry.MustRegister(collector.NewGageCollector(s, client, sites, logger))

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handleRoot))
	mux.Handle("/healthz", http.HandlerFunc(handleHealthz))
	mux.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	logger.Info("Starting",
		"endpoint", *endpoint,
		"metrics", *metricsPath,
	)
	logger.Error("unable to start server",
		"err", http.ListenAndServe(*endpoint, mux),
	)
}
