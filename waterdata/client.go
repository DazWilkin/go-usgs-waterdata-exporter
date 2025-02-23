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
	endpoint string = "https://waterservices.usgs.gov/nwis/iv/"
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
	Client *http.Client
	Logger *slog.Logger
}

// NewClient is a function that returns a new Client
func NewClient(l *slog.Logger) *Client {
	return &Client{
		Client: &http.Client{},
		Logger: l,
	}
}

// GetInstantaneousValues is a method that returns values using the InstantaneousValues Service
// https://waterservices.usgs.gov/nwis/iv/?format=json&sites=12150400&modifiedSince=PT1H&siteStatus=all
// Constants applied to HTTP requests
// format=json
// sites=site1,site2,...
// modifiedSince=PT1H (ISO 8601 Duration: See https://en.wikipedia.org/wiki/ISO_8601#Durations)
// siteStatus=all
func (c *Client) GetInstantaneousValues(sites []string) (*GetInstantaneousValuesResponse, error) {
	params := url.Values{}
	params.Add("format", "json")
	params.Add("sites", strings.Join(sites, ","))
	params.Add("modifiedSince", "PT1H")
	params.Add("siteStatus", "all")

	queryString := params.Encode()

	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		msg := "unable to parse URL"
		slog.Error(msg,
			"err", err,
		)
		return &GetInstantaneousValuesResponse{}, errors.New(msg)
	}

	parsedURL.RawQuery = queryString
	fullURL := parsedURL.String()

	rqst, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		msg := "unable to create request"
		slog.Error(msg,
			"err", err,
		)
		return &GetInstantaneousValuesResponse{}, errors.New(msg)
	}

	resp, err := c.Client.Do(rqst)
	if err != nil {
		msg := "unable to make request"
		slog.Error(msg,
			"err", err,
		)
		return &GetInstantaneousValuesResponse{}, errors.New(msg)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := "unable to read response body"
		slog.Error(msg,
			"err", err,
		)
		return &GetInstantaneousValuesResponse{}, errors.New(msg)
	}

	if resp.StatusCode != http.StatusOK {
		return &GetInstantaneousValuesResponse{}, errors.New(resp.Status)
	}

	respMsg := &GetInstantaneousValuesResponse{}
	if err := json.Unmarshal(respBody, respMsg); err != nil {
		msg := "unable to unmarshal response body"
		slog.Error(msg,
			"err", err,
		)
		return &GetInstantaneousValuesResponse{}, errors.New(msg)
	}

	return respMsg, nil
}
