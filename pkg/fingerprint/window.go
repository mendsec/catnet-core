package fingerprint

// GuessOSFromTCPParams attempts to detect the OS and device type using both TTL and TCP Window Size parameters.
// If windowSize is 0, heuristic analysis falls back exclusively to TTL-based detection.
func GuessOSFromTCPParams(ttl int, windowSize int) FingerprintResult {
	res := FingerprintResult{
		OS:         "Unknown",
		OSFamily:   "unknown",
		DeviceType: DeviceUnknown,
		AssetClass: AssetClassUnknown,
		Confidence: 0,
	}

	if ttl <= 0 {
		return res
	}

	// Router / Low TTL path
	if ttl < 30 {
		res.DeviceType = DeviceRouter
		res.AssetClass = AssetClassNetwork
		res.Confidence = 50
		return res
	}

	// TTL around 64 (Linux, macOS, BSD, Android, iOS)
	if ttl >= 60 && ttl <= 64 {
		res.OSFamily = "linux"
		res.OS = "Linux/Unix"
		res.DeviceType = DeviceServer
		res.AssetClass = AssetClassCompute
		res.Confidence = 75

		// TCP Window Size refinement for TTL 64
		switch windowSize {
		case 5840, 14600, 29200, 64240:
			res.OSFamily = "linux"
			res.OS = "Linux"
			res.Confidence = 85
		case 16384, 32768, 65535:
			// Common default windows for FreeBSD, macOS, iOS
			res.OSFamily = "unix"
			res.OS = "macOS/BSD"
			res.Confidence = 80
		}
		return res
	}

	// TTL around 128 (Windows)
	if ttl >= 120 && ttl <= 128 {
		res.OSFamily = "windows"
		res.OS = "Windows"
		res.DeviceType = DeviceWorkstation
		res.AssetClass = AssetClassCompute
		res.Confidence = 80

		// TCP Window Size refinement for TTL 128
		switch windowSize {
		case 8192, 64512, 65535:
			res.Confidence = 90
		}
		return res
	}

	// TTL around 255 (Cisco, Solaris, BSD, Network Switches)
	if ttl >= 250 && ttl <= 255 {
		res.OSFamily = "unix"
		res.OS = "Cisco/Unix"
		res.DeviceType = DeviceRouter
		res.AssetClass = AssetClassNetwork
		res.Confidence = 65

		switch windowSize {
		case 4128, 8760:
			res.OS = "Cisco IOS"
			res.Confidence = 85
		}
		return res
	}

	return res
}
