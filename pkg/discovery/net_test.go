package discovery

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// These tests verify input validation paths only. Network-dependent paths are covered by integration tests.

func TestPingValidation(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{"empty string", ""},
		{"invalid ip", "999.999.999.999"},
		{"not an ip", "not-an-ip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if Ping(context.Background(), tt.ip, 1000) {
				t.Errorf("Ping(%q) expected false, got true", tt.ip)
			}
		})
	}
}

func TestReverseDNSValidation(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{"empty string", ""},
		{"out of bounds ip", "256.0.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if res := ReverseDNS(context.Background(), tt.ip); res != "" {
				t.Errorf("ReverseDNS(%q) expected empty string, got %q", tt.ip, res)
			}
		})
	}
}

func TestGetMACValidation(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{"empty string", ""},
		{"ipv6 not supported", "::1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if res := GetMAC(context.Background(), tt.ip); res != "" {
				t.Errorf("GetMAC(%q) expected empty string, got %q", tt.ip, res)
			}
		})
	}
}

func TestReverseDNSCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res := ReverseDNS(ctx, "8.8.8.8")
	if res != "" {
		t.Errorf("ReverseDNS() with cancelled context should return empty string, got: %q", res)
	}
}

func TestGetMACCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res := GetMAC(ctx, "127.0.0.1")
	if res != "" {
		t.Errorf("GetMAC() with cancelled context should return empty string, got: %q", res)
	}
}

func TestPingCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	before := runtime.NumGoroutine()

	res := Ping(ctx, "127.0.0.1", 1000)
	if res {
		t.Errorf("Ping() with cancelled context should return false, got: %v", res)
	}

	after := runtime.NumGoroutine()
	if after > before {
		t.Errorf("Goroutine leak detected in Ping with cancelled context: before=%d, after=%d", before, after)
	}
}

func TestPingCancellationActive(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	before := runtime.NumGoroutine()

	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	res := Ping(ctx, "192.0.2.1", 2000)
	if res {
		t.Errorf("Ping() with actively cancelled context should return false, got: %v", res)
	}

	for i := 0; i < 10; i++ {
		if runtime.NumGoroutine() <= before+1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	after := runtime.NumGoroutine()
	if after > before+1 {
		t.Errorf("Goroutine leak detected in active Ping cancellation: before=%d, after=%d", before, after)
	}
}
