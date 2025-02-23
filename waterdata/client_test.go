package waterdata

import (
	"log/slog"
	"os"
	"testing"
)

// TestInstantaneousValues tests whether the service returns a value
// TODO Improve the test to check the value that's returned
func TestInstantaneousValues(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	client := NewClient(logger)

	sites := []string{
		SnoqualmieCarnation,
		SnoqualmieDuvall,
	}
	resp, err := client.GetInstantaneousValues(sites)
	if err != nil {
		t.Errorf("expected success")
	}

	print(resp)
}
