package collector

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/DazWilkin/go-probe/probe"

	"github.com/DazWilkin/go-usgs-waterdata/waterdata"
)

var (
	// waterdataResponse string = `
	// {
	// 	"value": {
	// 		"timeSeries": [
	// 			{
	// 				"sourceInfo": {
	// 					"siteCode": [
	// 						{
	// 							"value": "12149000"
	// 						}
	// 					]
	// 				},
	// 				"variable": {
	// 					"variableCode": [
	// 						{
	// 							"value": "00060"
	// 						}
	// 					]
	// 				},
	// 				"values": [
	// 					{
	// 						"value": [
	// 							{
	// 								"value": "52.17"
	// 							}
	// 						]
	// 					}
	// 				]
	// 			}
	// 		]
	// 	}
	// }`
	waterdataResponse string = func() string {
		data, err := os.ReadFile("../examples/instantaneous_values.json")
		if err != nil {
			panic(err)
		}
		return string(data)
	}()
	prometheusResponse string = `
	# HELP iv_gage_height_feet Gage Height Feet
	# TYPE iv_gage_height_feet gauge
	iv_gage_height_feet{site="12149000"} 52.17
	iv_gage_height_feet{site="12150400"} 34.18
	`
)

func TestInstantaneousValuesCollector(t *testing.T) {
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	p := probe.New("test", l)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan probe.Status)
	go p.Updater(ctx, ch, nil)

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, waterdataResponse)
	})

	s := System{}
	c, err := waterdata.NewClient(l)
	if err != nil {
		t.Errorf("expected to be able to create client")
	}

	c.BaseURL, err = url.Parse(server.URL)
	if err != nil {
		t.Errorf("expected to be able to override client's base URL")
	}

	sitecodes := []string{
		"12149000",
		"12150400",
	}

	collector := NewInstantaneousValuesCollector(s, c, ch, sitecodes, l)
	if err := testutil.CollectAndCompare(
		collector,
		strings.NewReader(prometheusResponse),
	); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}
