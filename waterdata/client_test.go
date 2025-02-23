package waterdata

import (
	"log/slog"
	"os"
	"testing"
)

// TestGetGauge tests whether GetGage returns a value
// TODO Improve the test to check the value that's returned
func TestGetGauge(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	client := NewClient(logger)

	sites := []string{
		SnoqualmieCarnation,
		SnoqualmieDuvall,
	}
	resp, err := client.GetGage(sites)
	if err != nil {
		t.Errorf("expected success")
	}

	print(resp)
}
