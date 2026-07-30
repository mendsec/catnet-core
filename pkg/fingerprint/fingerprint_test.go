package fingerprint

import (
	"context"
	"testing"
)

func TestGuessOSFromTCPParams(t *testing.T) {
	tests := []struct {
		name       string
		ttl        int
		window     int
		osFamily   string
		deviceType DeviceType
		assetClass AssetClass
		minConf    int
	}{
		{
			name:       "Linux with standard TTL 64 and TCP window 5840",
			ttl:        64,
			window:     5840,
			osFamily:   "linux",
			deviceType: DeviceServer,
			assetClass: AssetClassCompute,
			minConf:    85,
		},
		{
			name:       "macOS/BSD with TTL 64 and TCP window 65535",
			ttl:        64,
			window:     65535,
			osFamily:   "unix",
			deviceType: DeviceServer,
			assetClass: AssetClassCompute,
			minConf:    80,
		},
		{
			name:       "Windows with TTL 128 and TCP window 8192",
			ttl:        128,
			window:     8192,
			osFamily:   "windows",
			deviceType: DeviceWorkstation,
			assetClass: AssetClassCompute,
			minConf:    90,
		},
		{
			name:       "Cisco IOS with TTL 255 and TCP window 4128",
			ttl:        255,
			window:     4128,
			osFamily:   "unix",
			deviceType: DeviceRouter,
			assetClass: AssetClassNetwork,
			minConf:    85,
		},
		{
			name:       "Low TTL router fallback",
			ttl:        15,
			window:     0,
			osFamily:   "unknown",
			deviceType: DeviceRouter,
			assetClass: AssetClassNetwork,
			minConf:    50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := GuessOSFromTCPParams(tt.ttl, tt.window)
			if res.OSFamily != tt.osFamily {
				t.Errorf("expected OSFamily %s, got %s", tt.osFamily, res.OSFamily)
			}
			if res.DeviceType != tt.deviceType {
				t.Errorf("expected DeviceType %s, got %s", tt.deviceType, res.DeviceType)
			}
			if res.AssetClass != tt.assetClass {
				t.Errorf("expected AssetClass %s, got %s", tt.assetClass, res.AssetClass)
			}
			if res.Confidence < tt.minConf {
				t.Errorf("expected Confidence >= %d, got %d", tt.minConf, res.Confidence)
			}
		})
	}
}

func TestOsFromBanners(t *testing.T) {
	tests := []struct {
		name       string
		banners    map[int]string
		os         string
		osFamily   string
		deviceType DeviceType
		assetClass AssetClass
	}{
		{
			name:       "Ubuntu SSH banner",
			banners:    map[int]string{22: "SSH-2.0-OpenSSH_8.9p1 Ubuntu"},
			os:         "Ubuntu",
			osFamily:   "linux",
			deviceType: DeviceServer,
			assetClass: AssetClassCompute,
		},
		{
			name:       "Windows cross-banner (SMB + RDP + IIS)",
			banners:    map[int]string{445: "SMB", 3389: "MS-RDP", 80: "Server: IIS/10.0"},
			os:         "Windows",
			osFamily:   "windows",
			deviceType: DeviceServer,
			assetClass: AssetClassCompute,
		},
		{
			name:       "Printer ports (9100 JetDirect)",
			banners:    map[int]string{9100: "JetDirect"},
			os:         "Embedded Printer",
			osFamily:   "unix",
			deviceType: DeviceIoT,
			assetClass: AssetClassIoT,
		},
		{
			name:       "NAS Storage (Synology DSM)",
			banners:    map[int]string{5000: "HTTP/1.1 200 OK (Synology DSM)"},
			os:         "Synology DSM",
			osFamily:   "linux",
			deviceType: DeviceServer,
			assetClass: AssetClassStorage,
		},
		{
			name:       "MikroTik RouterOS",
			banners:    map[int]string{8291: "Winbox / MikroTik RouterOS"},
			os:         "MikroTik RouterOS",
			osFamily:   "linux",
			deviceType: DeviceRouter,
			assetClass: AssetClassNetwork,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := OsFromBanners(tt.banners)
			if res.OSFamily != tt.osFamily {
				t.Errorf("expected OSFamily %s, got %s", tt.osFamily, res.OSFamily)
			}
			if res.OS != tt.os {
				t.Errorf("expected OS %s, got %s", tt.os, res.OS)
			}
			if res.DeviceType != tt.deviceType {
				t.Errorf("expected DeviceType %s, got %s", tt.deviceType, res.DeviceType)
			}
			if res.AssetClass != tt.assetClass {
				t.Errorf("expected AssetClass %s, got %s", tt.assetClass, res.AssetClass)
			}
		})
	}
}

func TestFingerprint(t *testing.T) {
	ctx := context.Background()
	res := Fingerprint(ctx, "127.0.0.1", "B8:27:EB:11:22:33", 64, []int{}, 100)
	if res.Vendor != "Raspberry Pi" {
		t.Errorf("expected vendor Raspberry Pi, got %s", res.Vendor)
	}
	if res.DeviceType != DeviceIoT {
		t.Errorf("expected DeviceType %s, got %s", DeviceIoT, res.DeviceType)
	}
	if res.AssetClass != AssetClassIoT {
		t.Errorf("expected AssetClass %s, got %s", AssetClassIoT, res.AssetClass)
	}
}

func TestFingerprintWithParams(t *testing.T) {
	ctx := context.Background()
	cfg := BannerGrabConfig{
		AggressiveSMB: false,
		Concurrency:   5,
	}
	res := FingerprintWithParams(ctx, "127.0.0.1", "", 128, 8192, []int{}, 100, cfg)
	if res.OSFamily != "windows" {
		t.Errorf("expected OSFamily windows, got %s", res.OSFamily)
	}
	if res.AssetClass != AssetClassCompute {
		t.Errorf("expected AssetClass compute, got %s", res.AssetClass)
	}
}
