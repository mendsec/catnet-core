package fingerprint

// GuessOSFromTTL attempts to detect the OS and device type based on the IP TTL value.
func GuessOSFromTTL(ttl int) FingerprintResult {
	return GuessOSFromTCPParams(ttl, 0)
}
