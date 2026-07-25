package targets

import (
	"errors"
	"net"
	"testing"

	"github.com/catnet-io/engine/pkg/coreerr"
)

func TestParseRange(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantErr bool
	}{
		{"Single IP valid", "10.0.0.1", 1, false},
		{"CIDR valid", "192.168.1.0/30", 2, false}, // 192.168.1.1 and 192.168.1.2
		{"CIDR too large", "0.0.0.0/0", 0, true},
		{"CIDR max limit", "10.0.0.0/16", 65534, false},
		{"Dash range full valid", "192.168.1.10-192.168.1.12", 3, false},
		{"Dash range shorthand valid", "192.168.1.10-12", 3, false},
		{"Invalid format", "invalid", 0, true},
		{"Empty input", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRange(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("ParseRange() got %v items, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestParseRangeAuto(t *testing.T) {
	for _, input := range []string{"auto", "AUTO", " Auto "} {
		t.Run(input, func(t *testing.T) {
			got, err := ParseRange(input)
			if err != nil {
				// If local host has no active network interfaces, error wrapping coreerr.ErrInvalidInput is expected
				if !errors.Is(err, coreerr.ErrInvalidInput) {
					t.Fatalf("ParseRange(%q) returned unexpected error: %v", input, err)
				}
				return
			}
			if len(got) == 0 {
				t.Fatalf("ParseRange(%q) returned empty slice without error", input)
			}
		})
	}
}

func TestDetectLocalRangeLive(t *testing.T) {
	ips, err := DetectLocalRange()
	if err != nil {
		if errors.Is(err, coreerr.ErrInvalidInput) {
			t.Skipf("Skipping live DetectLocalRange test: no active non-loopback network interface detected on host: %v", err)
		}
		t.Fatalf("DetectLocalRange() unexpected error: %v", err)
	}
	if len(ips) == 0 {
		t.Fatalf("DetectLocalRange() returned empty IP list")
	}
}

func TestDetectLocalRangeCustom(t *testing.T) {
	_, ipnetValid1, _ := net.ParseCIDR("192.168.1.1/30")
	_, ipnetValid2, _ := net.ParseCIDR("192.168.1.2/30")
	_, ipnetLinkLocal, _ := net.ParseCIDR("169.254.1.1/16")
	_, ipnetV6, _ := net.ParseCIDR("fe80::1/64")

	t.Run("Valid interfaces and deduplication", func(t *testing.T) {
		getIfaces := func() ([]net.Interface, error) {
			return []net.Interface{
				{Index: 1, Name: "lo", Flags: net.FlagUp | net.FlagLoopback},
				{Index: 2, Name: "eth0_down", Flags: 0},
				{Index: 3, Name: "eth0", Flags: net.FlagUp | net.FlagBroadcast},
				{Index: 4, Name: "eth1", Flags: net.FlagUp | net.FlagBroadcast},
			}, nil
		}

		getAddrs := func(iface net.Interface) ([]net.Addr, error) {
			switch iface.Name {
			case "eth0":
				return []net.Addr{ipnetValid1, ipnetLinkLocal, ipnetV6}, nil
			case "eth1":
				return []net.Addr{ipnetValid2}, nil
			default:
				return nil, nil
			}
		}

		got, err := detectLocalRangeCustom(getIfaces, getAddrs)
		if err != nil {
			t.Fatalf("detectLocalRangeCustom unexpected error: %v", err)
		}

		// 192.168.1.0/30 expands to 192.168.1.1 and 192.168.1.2.
		// Both eth0 and eth1 cover 192.168.1.0/30, so deduplication must result in exactly 2 IPs.
		if len(got) != 2 {
			t.Fatalf("expected 2 deduplicated IPs, got %d: %v", len(got), got)
		}
		expected := []string{"192.168.1.1", "192.168.1.2"}
		for i, ip := range got {
			if ip != expected[i] {
				t.Errorf("got[%d] = %s, want %s", i, ip, expected[i])
			}
		}
	})

	t.Run("No active interfaces error", func(t *testing.T) {
		getIfaces := func() ([]net.Interface, error) {
			return []net.Interface{
				{Index: 1, Name: "lo", Flags: net.FlagUp | net.FlagLoopback},
			}, nil
		}
		getAddrs := func(iface net.Interface) ([]net.Addr, error) {
			return nil, nil
		}

		_, err := detectLocalRangeCustom(getIfaces, getAddrs)
		if err == nil {
			t.Fatalf("expected error when no active non-loopback interface exists, got nil")
		}
		if !errors.Is(err, coreerr.ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got: %v", err)
		}
	})

	t.Run("Interface fetch failure", func(t *testing.T) {
		getIfaces := func() ([]net.Interface, error) {
			return nil, errors.New("network error")
		}
		getAddrs := func(iface net.Interface) ([]net.Addr, error) {
			return nil, nil
		}

		_, err := detectLocalRangeCustom(getIfaces, getAddrs)
		if err == nil {
			t.Fatalf("expected error when getIfaces fails, got nil")
		}
	})
}

func BenchmarkParseRange(b *testing.B) {
	input := "10.0.0.0/16" // 65534 IPs
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseRange(input)
	}
}

func FuzzParseRange(f *testing.F) {
	f.Add("10.0.0.1")
	f.Add("192.168.1.0/24")
	f.Add("192.168.1.1-192.168.1.10")
	f.Add("10.0.0.1-255")
	f.Add("auto")
	f.Add("AUTO")
	f.Add("invalid")
	f.Add("10.0.0.0/0")
	f.Add("10.0.0.0/33")

	f.Fuzz(func(t *testing.T, input string) {
		// The goal of fuzzing here is to ensure ParseRange never panics on malformed input.
		_, _ = ParseRange(input)
	})
}

