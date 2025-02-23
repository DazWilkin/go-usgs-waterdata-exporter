package collector

import (
	"log/slog"
	"strconv"

	"github.com/DazWilkin/go-usgs-waterdata/waterdata"

	"github.com/prometheus/client_golang/prometheus"
)

// GageCollector represents the result of USGS Waterdata Gage measurements
// Currently this only extracts one measurement GageHeightFeet ("00065")
type GageCollector struct {
	System System
	Client *waterdata.Client
	Logger *slog.Logger

	Sites []string

	GageHeightFeet *prometheus.Desc
}

// NewGageCollector is a function that creates a new GageCollector
func NewGageCollector(s System, c *waterdata.Client, sites []string, l *slog.Logger) *GageCollector {
	subsystem := "gage"
	return &GageCollector{
		System: s,
		Client: c,
		Logger: l,

		Sites: sites,

		GageHeightFeet: prometheus.NewDesc(
			prometheus.BuildFQName(s.Namespace, subsystem, "height_feet"),
			"Gage Height Feet",
			[]string{
				"site",
			},
			nil,
		),
	}
}

// Collect is a method that implements Prometheus' Collector interface and collects metrics
func (c *GageCollector) Collect(ch chan<- prometheus.Metric) {
	resp, err := c.Client.GetGage(c.Sites)
	if err != nil {
		slog.Info("unable to get waterdata gage")
		return
	}

	// []TimeSeries contains more than just Gage Height measurements
	// Must filter results to ensure the VariableCode[].Value only contains GageHeightFeet ("00065")
	// TODO extend this filter if other Prometheus Metrics are added
	for _, t := range resp.Value.TimeSeries {
		// Filter TimeSeries to only those where VariableCode[].Value contains GageHeightFeet
		if !func(vv []waterdata.VariableCode, s string) bool {
			for _, v := range vv {
				if v.Value == s {
					return true
				}
			}

			return false
		}(t.Variable.VariableCode, waterdata.GageHeightFeet) {
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
					slog.Info("unable to parse value as float64",
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
func (c *GageCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.GageHeightFeet
}
