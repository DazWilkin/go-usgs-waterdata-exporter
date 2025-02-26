package collector

import (
	"errors"
	"fmt"
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

	rqstMap := make(map[string]bool)

	// Mark all requested sites as existing
	for _, site := range c.Sites {
		rqstMap[site] = true
	}

	respMap := make(map[string]bool)

	// []TimeSeries contains more than just Gage Height measurements
	// Must filter results to ensure the VariableCode[].Value only contains GageHeightFeet ("00065")
	// TODO extend this filter if other Prometheus Metrics are added
	for _, t := range resp.Value.TimeSeries {
		logger.Info("Iterating over results",
			"name", t.Name,
		)

		sitecode, err := c.getSitecode(t)
		if err != nil {
			logger.Info("Unable to get sitecode",
				"name", t.Name,
				"error", err,
			)
			continue
		}

		// Compare requested c.Sites and returned resp.Value.TimeSeries.SourceInfo.Sitecode
		respMap[sitecode] = true

		// Filter TimeSeries to only those where VariableCode[].Value contains GageHeightFeet
		if !t.Variable.Contains(waterdata.GageHeightFeet) {
			logger.Info("Excluded",
				"name", t.Name,
			)
			continue
		}

		value, err := c.getValue(t)
		if err != nil {
			logger.Info("Unable to get value",
				"name", t.Name,
				"error", err,
			)
			continue
		}
		logger.Info("Measured",
			"name", t.Name,
			"sitecode", sitecode,
			"value", value,
		)

		// Only GageHeightFeet measurements are left
		// TODO Check to corrorboate sites parameter with values returned

		ch <- prometheus.MustNewConstMetric(
			c.GageHeightFeet,
			prometheus.GaugeValue,
			value,
			[]string{
				sitecode,
			}...,
		)
	}

	if notResp, notRqst := c.findMismatches(rqstMap, respMap); len(notResp) != 0 || len(notRqst) != 0 {
		logger.Info("Sites not matched",
			"requested but not responded", notResp,
			"responded but not requested", notRqst,
		)
	}
}

// Describe is a method that implements Prometheus' Collector interface and describes metrics
func (c *InstantaneousValuesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.GageHeightFeet
}

// getSitecode is a method that returns the sitecode from the TimeSeries
func (c *InstantaneousValuesCollector) getSitecode(t waterdata.TimeSeries) (string, error) {
	logger := c.Logger.With("method", "getSitecode")

	if len(t.SourceInfo.SiteCode) == 0 {
		msg := "no site code returned"
		logger.Info(msg,
			"name", t.Name,
		)
		return "", errors.New(msg)
	}

	logger.Info("Sitecodes returned",
		"name", t.Name,
		"sitecodes", len(t.SourceInfo.SiteCode),
	)

	sitecode := t.SourceInfo.SiteCode[0].Value

	return sitecode, nil
}

// getValue is a method that returns the value from the TimeSeries
// It only returns the first value in the Values array
func (c *InstantaneousValuesCollector) getValue(t waterdata.TimeSeries) (float64, error) {
	logger := c.Logger.With("method", "getValue")

	if len(t.Values) == 0 {
		msg := "no values returned"
		logger.Info(msg,
			"name", t.Name,
		)
		return 0.0, errors.New(msg)
	}

	logger.Info("Values returned",
		"name", t.Name,
		"values", len(t.Values),
	)

	if len(t.Values[0].Value) == 0 {
		msg := "no Value[0] values returned"
		logger.Info(msg,
			"name", t.Name,
			"values", len(t.Values[0].Value),
		)
		return 0.0, errors.New(msg)
	}

	logger.Info("Values[0] values (!) returned",
		"name", t.Name,
		"values", len(t.Values[0].Value),
	)

	v := t.Values[0].Value[0].Value
	value, err := strconv.ParseFloat(v, 64)
	if err != nil {
		msg := fmt.Sprintf("unable to parse value (%s) as float64", v)
		logger.Info(msg,
			"value", v,
		)
		return 0.0, errors.New(msg)
	}

	return value, nil

}

// findMismatches is a method that returns the sites that are not matched
// It returns two slices, one for sites that were requested but not responded
// and one for sites that were responded but not requested
func (c *InstantaneousValuesCollector) findMismatches(rqstMap, respMap map[string]bool) ([]string, []string) {
	// logger := c.Logger.With("method", "findMismatches")

	// Iterate over maps determining which sites are missing
	// Requested but not responded
	notResp := []string{}
	for site := range rqstMap {
		if _, ok := respMap[site]; !ok {
			notResp = append(notResp, site)
		}
	}

	// Responded but not requested
	notRqst := []string{}
	for site := range respMap {
		if _, ok := rqstMap[site]; !ok {
			notRqst = append(notRqst, site)
		}
	}

	return notResp, notRqst
}
