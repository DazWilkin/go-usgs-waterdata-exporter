package collector

import (
	"log/slog"
	"strconv"

	"github.com/DazWilkin/go-usgs-waterdata-exporter/waterdata"

	"github.com/prometheus/client_golang/prometheus"
)

// InstantaneousValuesCollector represents the result of USGS Instantaneous Values service
// Currently this only extracts one measurement GageHeightFeet ("00065")
type InstantaneousValuesCollector struct {
	System System
	Client *waterdata.Client
	Logger *slog.Logger

	Sites []string

	GageHeightFeet *prometheus.Desc
}

// NewInstantaneousValuesCollector is a function that creates a new GageCollector
func NewInstantaneousValuesCollector(s System, c *waterdata.Client, sites []string, l *slog.Logger) *InstantaneousValuesCollector {
	subsystem := "iv"
	logger := l.With("collector", subsystem)
	return &InstantaneousValuesCollector{
		System: s,
		Client: c,
		Logger: logger,

		Sites: sites,

		GageHeightFeet: prometheus.NewDesc(
			prometheus.BuildFQName(s.Namespace, subsystem, "gage_height_feet"),
			"Gage Height Feet",
			[]string{
				"site",
			},
			nil,
		),
	}
}

// Collect is a method that implements Prometheus' Collector interface and collects metrics
func (c *InstantaneousValuesCollector) Collect(ch chan<- prometheus.Metric) {
	logger := c.Logger.With("method", "collect")
	resp, err := c.Client.GetInstantaneousValues(c.Sites)
	if err != nil {
		logger.Info("Unable to get waterdata gage")
		return
	}

	// []TimeSeries contains more than just Gage Height measurements
	// Must filter results to ensure the VariableCode[].Value only contains GageHeightFeet ("00065")
	// TODO extend this filter if other Prometheus Metrics are added
	for _, t := range resp.Value.TimeSeries {
		logger.Info("Iterating over results",
			"name", t.Name,
		)

		// Filter TimeSeries to only those where VariableCode[].Value contains GageHeightFeet
		if !t.Variable.Contains(waterdata.GageHeightFeet) {
			logger.Info("Excluded",
				"name", t.Name,
			)
			continue
		}

		// Only GageHeightFeet measurements are left
		// TODO Check to corrorboate sites parameter with values returned
		ch <- prometheus.MustNewConstMetric(
			c.GageHeightFeet,
			prometheus.GaugeValue,
			func(v string) float64 {
				r, err := strconv.ParseFloat(v, 64)
				if err != nil {
					logger.Info("Unable to parse value as float64",
						"value", v,
					)
					return 0.0
				}

				return r
			}(t.Values[0].Value[0].Value),
			[]string{
				//TODO Why is it a slice of SiteCode?
				t.SourceInfo.SiteCode[0].Value,
			}...,
		)
	}
}

// Describe is a method that implements Prometheus' Collector interface and describes metrics
func (c *InstantaneousValuesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.GageHeightFeet
}
