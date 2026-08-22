// Package fingerprint provides heuristic and banner-based operating system
// and device type detection mechanisms.
//
// Main exports:
// - Fingerprint: Orchestrates OS, device type, and vendor detection.
// - GrabBanners: Collects banners of open ports via TCP connection.
// - GuessOSFromTTL: Detects OS family from TTL value.
// - VendorFromMAC: Identifies the manufacturer from the MAC OUI prefix.
// - OsFromBanners: Infers OS and device type from collected banners.
package fingerprint
