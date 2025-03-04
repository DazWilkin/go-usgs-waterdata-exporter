package waterdata

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

const (
	// iv = Instantaneous Values service
	rawURL string = "https://waterservices.usgs.gov/nwis/iv/"
)
const (
	// Refactor this
	SnoqualmieCarnation string = "12149000"
	SnoqualmieDuvall    string = "12150400"
)
const (
	// Corresponds to TimeSeries.Variable.VariableCode[?].Value
	GageHeightFeet string = "00065"
)

// Client is a type that represents a Waterdata Service client
type Client struct {
	// Allow the URL to be overridden for testing
	BaseURL *url.URL
	Client  *http.Client
	Logger  *slog.Logger
}

// NewClient is a function that returns a new Client
func NewClient(l *slog.Logger) (*Client, error) {
	logger := l.With("client", "waterdata")

	baseURL, err := url.Parse(rawURL)
	if err != nil {
		msg := "unable to parse URL"
		logger.Error(msg,
			"endpoint", rawURL,
			"err", err,
		)
		return &Client{}, errors.New(msg)
	}

	return &Client{
		BaseURL: baseURL,
		Client:  &http.Client{},
		Logger:  logger,
	}, nil
}

// GetInstantaneousValues is a method that returns values using the InstantaneousValues Service
// https://waterservices.usgs.gov/nwis/iv/?format=json&sites=12150400&modifiedSince=PT1H&siteStatus=all
// Constants applied to HTTP requests
// format=json
// sites=site1,site2,...
// modifiedSince=PT1H (ISO 8601 Duration: See https://en.wikipedia.org/wiki/ISO_8601#Durations)
// siteStatus=all
func (c *Client) GetInstantaneousValues(sites []string) (*GetInstantaneousValuesResponse, error) {
	logger := c.Logger
	logger.Info("Get",
		"sites", sites,
	)

	params := url.Values{}
	params.Add("format", "json")
	params.Add("sites", strings.Join(sites, ","))
	params.Add("modifiedSince", "PT1H")
	params.Add("siteStatus", "active")

	queryString := params.Encode()

	c.BaseURL.RawQuery = queryString
	fullURL := c.BaseURL.String()

	rqst, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		msg := "unable to create request"
		logger.Error(msg,
			"err", err,
		)
		return &GetInstantaneousValuesResponse{}, errors.New(msg)
	}

	logger.Info("Invoking method")
	resp, err := c.Client.Do(rqst)
	if err != nil {
		msg := "unable to make request"
		logger.Error(msg,
			"err", err,
		)
		return &GetInstantaneousValuesResponse{}, errors.New(msg)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := "unable to read response body"
		logger.Error(msg,
			"err", err,
		)
		return &GetInstantaneousValuesResponse{}, errors.New(msg)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("non-200 response")
		return &GetInstantaneousValuesResponse{}, errors.New(resp.Status)
	}

	logger.Info("Unmarshaling response body")
	respMsg := &GetInstantaneousValuesResponse{}
	if err := json.Unmarshal(respBody, respMsg); err != nil {
		msg := "unable to unmarshal response body"
		logger.Error(msg,
			"err", err,
		)
		return &GetInstantaneousValuesResponse{}, errors.New(msg)
	}

	logger.Info("Returning response",
		"timeseries", len(respMsg.Value.TimeSeries),
	)
	return respMsg, nil
}
