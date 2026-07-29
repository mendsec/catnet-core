package results_test

import (
	"testing"

	"github.com/catnet-io/engine/pkg/results"
)

func TestWellKnownServiceName(t *testing.T) {
	tests := []struct {
		port     int
		expected string
	}{
		{22, "SSH"},
		{80, "HTTP"},
		{443, "HTTPS"},
		{3389, "RDP"},
		{9999, ""},
	}

	for _, tt := range tests {
		got := results.WellKnownServiceName(tt.port)
		if got != tt.expected {
			t.Errorf("results.WellKnownServiceName(%d) = %q; want %q", tt.port, got, tt.expected)
		}
	}
}
