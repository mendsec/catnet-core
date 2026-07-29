package ports_test

import (
	"testing"

	"github.com/catnet-io/engine/pkg/ports"
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
		got := ports.WellKnownServiceName(tt.port)
		if got != tt.expected {
			t.Errorf("ports.WellKnownServiceName(%d) = %q; want %q", tt.port, got, tt.expected)
		}
	}
}

func FuzzWellKnownServiceName(f *testing.F) {
	f.Add(22)
	f.Add(80)
	f.Add(-1)

	f.Fuzz(func(t *testing.T, port int) {
		name := ports.WellKnownServiceName(port)
		if port <= 0 || port > 65535 {
			if name != "" {
				t.Errorf("expected empty string for out-of-range port %d, got %q", port, name)
			}
		}
	})
}

func BenchmarkWellKnownServiceName(b *testing.B) {
	testPorts := []int{22, 80, 443, 3389, 8080, 9999}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ports.WellKnownServiceName(testPorts[i%len(testPorts)])
	}
}
