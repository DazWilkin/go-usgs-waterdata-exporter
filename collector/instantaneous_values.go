package collector

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/DazWilkin/go-probe/probe"

	"github.com/DazWilkin/go-usgs-waterdata/waterdata"

	"github.com/prometheus/client_golang/prometheus"
)

// InstantaneousValuesCollector represents the result of USGS Instantaneous Values service
// Currently this only extracts one measurement GageHeightFeet ("00065")
type InstantaneousValuesCollector struct {
	system System
	client *waterdata.Client
	ch     chan<- probe.Status
	logger *slog.Logger

	Sitecodes []string

	GageHeightFeet *prometheus.Desc
}

// NewInstantaneousValuesCollector is a function that creates a new GageCollector
func NewInstantaneousValuesCollector(s System, c *waterdata.Client, ch chan<- probe.Status, sitecodes []string, l *slog.Logger) *InstantaneousValuesCollector {
	subsystem := "iv"
	logger := l.With("collector", subsystem)
	return &InstantaneousValuesCollector{
		system: s,
		client: c,
		ch:     ch,
		logger: logger,

		Sitecodes: sitecodes,

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
	logger := c.logger.With("method", "collect")
	resp, err := c.client.GetInstantaneousValues(c.Sitecodes)
	if err != nil {
		msg := "Unable to get waterdata gage"
		logger.Info(msg,
			"err", err.Error(),
		)

		// Send probe unhealthy status
		// Doesn't surface the API error message (should it!?)
		status := probe.Status{
			Healthy: false,
			Message: msg,
		}
		c.ch <- status

		return
	}

	logger.Info("Received response")

	// Send probe healthy status
	status := probe.Status{
		Healthy: true,
		Message: "ok",
	}
	c.ch <- status

	// Useful to be able to identify sitecodes that:
	// + requested but not responded (returned) by the service
	// + responded but not requested by the service
	// Use maps to track sites requested|responded
	rqstSitecodes := make(map[string]bool)
	respSitecodes := make(map[string]bool)

	// Mark all requested sites as existing
	for _, site := range c.Sitecodes {
		rqstSitecodes[site] = true
	}

	// []TimeSeries contains more than just Gage Height measurements
	// Must filter results to ensure the VariableCode[].Value only contains GageHeightFeet ("00065")
	// TODO extend this filter if other Prometheus Metrics are added
	for _, t := range resp.Value.TimeSeries {
		logger.Info("Iterating over results",
			"name", t.Name,
		)

		// Retrieve the Sitecode from the Timeseries
		sitecode, err := c.getSitecode(t)
		if err != nil {
			logger.Info("Unable to get sitecode",
				"name", t.Name,
				"error", err,
			)
			continue
		}

		// Track this sitecode as having been returned
		respSitecodes[sitecode] = true

		// Filter TimeSeries to only those where VariableCode[].Value contains GageHeightFeet
		if !t.Variable.Contains(waterdata.GageHeightFeet) {
			logger.Info("Excluded",
				"name", t.Name,
			)
			continue
		}

		// Retrieve the value from the Timeseries
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

		ch <- prometheus.MustNewConstMetric(
			c.GageHeightFeet,
			prometheus.GaugeValue,
			value,
			[]string{
				sitecode,
			}...,
		)
	}

	if notResp, notRqst := c.findMismatches(rqstSitecodes, respSitecodes); len(notResp) != 0 || len(notRqst) != 0 {
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
	logger := c.logger.With("method", "getSitecode")

	l := len(t.SourceInfo.SiteCode)

	if l == 0 {
		msg := "no site code returned"
		logger.Info(msg,
			"name", t.Name,
		)
		return "", errors.New(msg)
	}

	if l > 1 {
		logger.Info("Sitecodes returned",
			"name", t.Name,
			"sitecodes", l,
		)
	}

	sitecode := t.SourceInfo.SiteCode[0].Value

	return sitecode, nil
}

// getValue is a method that returns the value from the TimeSeries
// It only returns the first value in the Values array
func (c *InstantaneousValuesCollector) getValue(t waterdata.TimeSeries) (float64, error) {
	logger := c.logger.With("method", "getValue")

	// Check the length of the Values array
	{
		l := len(t.Values)

		if l == 0 {
			msg := "no values returned"
			logger.Info(msg,
				"name", t.Name,
			)
			return 0.0, errors.New(msg)
		}

		if l > 1 {
			logger.Info("Values returned",
				"name", t.Name,
				"values", l,
			)
		}
	}

	// Check the length of the Values[0].Value (!) array
	{
		l := len(t.Values[0].Value)
		if l == 0 {
			msg := "no Value[0] values returned"
			logger.Info(msg,
				"name", t.Name,
			)
			return 0.0, errors.New(msg)
		}

		if l > 1 {
			logger.Info("Values[0] values (!) returned",
				"name", t.Name,
				"values", l,
			)
		}
	}

	sValue := t.Values[0].Value[0].Value
	fValue, err := strconv.ParseFloat(sValue, 64)
	if err != nil {
		msg := fmt.Sprintf("unable to parse value (%s) as float64", sValue)
		logger.Info(msg,
			"value", sValue,
		)
		return 0.0, errors.New(msg)
	}

	return fValue, nil

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
