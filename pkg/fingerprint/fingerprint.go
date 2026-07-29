package fingerprint

import (
	"context"
)

// Fingerprint orchestrates the OS, device type, asset class, and vendor detection process.
func Fingerprint(ctx context.Context, ip, mac string, ttl int, openPorts []int, timeoutMs int) FingerprintResult {
	return FingerprintWithConfig(ctx, ip, mac, ttl, openPorts, timeoutMs, BannerGrabConfig{})
}

// FingerprintWithConfig is like Fingerprint but accepts a BannerGrabConfig.
func FingerprintWithConfig(ctx context.Context, ip, mac string, ttl int, openPorts []int, timeoutMs int, bc BannerGrabConfig) FingerprintResult {
	return FingerprintWithParams(ctx, ip, mac, ttl, 0, openPorts, timeoutMs, bc)
}

// FingerprintWithParams orchestrates detection with optional TCP window size and custom BannerGrabConfig.
func FingerprintWithParams(ctx context.Context, ip, mac string, ttl int, windowSize int, openPorts []int, timeoutMs int, bc BannerGrabConfig) FingerprintResult {
	// 1. Gather inputs
	tcpResult := GuessOSFromTCPParams(ttl, windowSize)

	banners := GrabBanners(ctx, ip, openPorts, timeoutMs, bc)
	bannerResult := OsFromBanners(banners)

	vendor := VendorFromMAC(mac)

	// 2. Combine logic based on priority
	var final FingerprintResult

	if bannerResult.Confidence > 70 {
		final = bannerResult
	} else if tcpResult.Confidence > 50 {
		final = tcpResult
	} else {
		final = bannerResult // Fallback
	}

	// 3. Set Vendor and refine DeviceType / AssetClass using known vendor signatures
	final.Vendor = vendor
	if vendor == "Raspberry Pi" || vendor == "Espressif" {
		final.DeviceType = DeviceIoT
		final.AssetClass = AssetClassIoT
	} else if vendor == "Ubiquiti" || vendor == "Cisco" || vendor == "Mikrotik" || vendor == "Arista" || vendor == "Juniper" {
		if final.DeviceType == DeviceUnknown || final.DeviceType == DeviceServer || final.DeviceType == "" {
			final.DeviceType = DeviceRouter
			final.AssetClass = AssetClassNetwork
		}
	}

	// 4. Ensure fields are initialized
	if final.OS == "" {
		final.OS = "Unknown"
	}
	if final.OSFamily == "" {
		final.OSFamily = "unknown"
	}
	if final.DeviceType == "" {
		final.DeviceType = DeviceUnknown
	}

	// 5. Compute AssetClass taxonomy if unassigned
	if final.AssetClass == "" || final.AssetClass == AssetClassUnknown {
		switch final.DeviceType {
		case DeviceRouter:
			final.AssetClass = AssetClassNetwork
		case DeviceServer, DeviceWorkstation:
			final.AssetClass = AssetClassCompute
		case DeviceIoT:
			final.AssetClass = AssetClassIoT
		case DeviceMobile:
			final.AssetClass = AssetClassMobile
		default:
			final.AssetClass = AssetClassUnknown
		}
	}

	return final
}
