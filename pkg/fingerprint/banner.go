package fingerprint

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// GrabBanners attempts to read banners from open ports.
func GrabBanners(ctx context.Context, ip string, openPorts []int, timeoutMs int, bc BannerGrabConfig) map[int]string {
	banners := make(map[int]string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	timeout := time.Duration(timeoutMs) * time.Millisecond

	workers := bc.Concurrency
	if workers <= 0 {
		workers = BannerConcurrency
	}
	if len(openPorts) < workers {
		workers = len(openPorts)
	}
	var index int32 = -1

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				idx := int(atomic.AddInt32(&index, 1))
				if idx >= len(openPorts) {
					return
				}
				banner := grabBannerFromPort(ctx, ip, openPorts[idx], timeout, bc)
				if banner != "" {
					mu.Lock()
					banners[openPorts[idx]] = banner
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	return banners
}

func grabBannerFromPort(ctx context.Context, ip string, port int, timeout time.Duration, bc BannerGrabConfig) string {
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return ""
	}
	defer conn.Close()

	stop := context.AfterFunc(ctx, func() {
		conn.Close()
	})
	defer stop()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	if port == 80 || port == 8080 {
		req := "HEAD / HTTP/1.0\r\n\r\n"
		_, _ = conn.Write([]byte(req))
	} else if port == 445 && bc.AggressiveSMB {
		// Basic SMB negotiate request — opt-in only; may trigger IDS/IPS alerts.
		smbReq := []byte{
			0x00, 0x00, 0x00, 0x2f, 0xff, 0x53, 0x4d, 0x42,
			0x72, 0x00, 0x00, 0x00, 0x00, 0x08, 0x01, 0x40,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0f, 0x0c,
			0x00, 0x00, 0x01, 0x00, 0x00, 0x0c, 0x00, 0x02,
			0x4e, 0x54, 0x20, 0x4c, 0x4d, 0x20, 0x30, 0x2e,
			0x31, 0x32, 0x00,
		}
		_, _ = conn.Write(smbReq)
	} else if port == 3389 {
		// RDP Connection Request (TPKT + X.224 Connection Request)
		rdpReq := []byte{
			0x03, 0x00, 0x00, 0x13, // TPKT Header (version 3, length 19)
			0x0e,                   // X.224 Length (14 bytes following)
			0xe0,                   // X.224 Connection Request PDTU
			0x00, 0x00,             // DST Reference
			0x00, 0x00,             // SRC Reference
			0x00,                   // Class & options
			0x01, 0x00, 0x08, 0x00, // RDP Neg Req: Type (0x01), Length (8)
			0x03, 0x00, 0x00, 0x00, // Protocols (0x03 = SSL + RDP)
		}
		_, _ = conn.Write(rdpReq)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err == nil && n > 0 {
		if port == 3389 && n >= 4 && buf[0] == 0x03 && buf[1] == 0x00 {
			return "MS-RDP"
		}
		rawBanner := string(buf[:n])
		validBanner := strings.ToValidUTF8(rawBanner, "?")
		return sanitizeBanner(validBanner)
	}
	return ""
}

func sanitizeBanner(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r >= 32 && r != 127 || r == '\t' || r == '\n' || r == '\r' {
			sb.WriteRune(r)
		}
	}
	return strings.TrimSpace(sb.String())
}

// OsFromBanners infers OS, device type, and asset class from collected banners.
func OsFromBanners(banners map[int]string) FingerprintResult {
	res := FingerprintResult{
		OS:         "Unknown",
		OSFamily:   "unknown",
		DeviceType: DeviceUnknown,
		AssetClass: AssetClassUnknown,
		Confidence: 0,
	}

	if len(banners) == 0 {
		return res
	}

	// 1. Check specialized asset classes first (Printers, NAS Storage, Routers)
	if printerRes, ok := isPrinterLike(banners); ok {
		return printerRes
	}

	if nasRes, ok := isNASLike(banners); ok {
		return nasRes
	}

	if routerRes, ok := isRouterLike(banners); ok {
		return routerRes
	}

	// 2. Cross-banner checks for OS platforms (Windows, Linux)
	if winRes, ok := isWindowsLike(banners); ok {
		return winRes
	}

	if linRes, ok := isLinuxLike(banners); ok {
		return linRes
	}

	// 3. Fallback generic banner parsing
	hasLighttpd := false
	_, has80 := banners[80]
	_, has8888 := banners[8888]
	_, has22 := banners[22]

	for port, banner := range banners {
		lowerBanner := strings.ToLower(banner)
		if strings.Contains(lowerBanner, "lighttpd") {
			hasLighttpd = true
		}

		if port == 80 || port == 8080 {
			if strings.Contains(lowerBanner, "apache") || strings.Contains(lowerBanner, "nginx") {
				res.OSFamily = "linux"
				res.OS = "Linux"
				res.DeviceType = DeviceServer
				res.AssetClass = AssetClassCompute
				res.Confidence = 80
			} else if strings.Contains(lowerBanner, "iis") {
				res.OSFamily = "windows"
				res.OS = "Windows"
				res.DeviceType = DeviceServer
				res.AssetClass = AssetClassCompute
				res.Confidence = 80
				return res
			}
		}
	}

	if res.DeviceType == DeviceUnknown {
		if hasLighttpd || ((has80 || has8888) && !has22) {
			res.DeviceType = DeviceIoT
			res.AssetClass = AssetClassIoT
			res.Confidence = 60
		}
	}

	return res
}

func isPrinterLike(banners map[int]string) (FingerprintResult, bool) {
	// JetDirect (9100), IPP (631), LPD (515)
	has9100 := false
	has631 := false
	has515 := false

	for port := range banners {
		if port == 9100 {
			has9100 = true
		} else if port == 631 {
			has631 = true
		} else if port == 515 {
			has515 = true
		}
	}

	if has9100 || (has631 && has515) {
		return FingerprintResult{
			OS:         "Embedded Printer",
			OSFamily:   "unix",
			DeviceType: DeviceIoT,
			AssetClass: AssetClassIoT,
			Confidence: 85,
		}, true
	}
	return FingerprintResult{}, false
}

func isNASLike(banners map[int]string) (FingerprintResult, bool) {
	for port, b := range banners {
		lower := strings.ToLower(b)
		if strings.Contains(lower, "synology") || port == 5000 || port == 5001 {
			return FingerprintResult{
				OS:         "Synology DSM",
				OSFamily:   "linux",
				DeviceType: DeviceServer,
				AssetClass: AssetClassStorage,
				Confidence: 90,
			}, true
		}
		if strings.Contains(lower, "qnap") {
			return FingerprintResult{
				OS:         "QNAP QTS",
				OSFamily:   "linux",
				DeviceType: DeviceServer,
				AssetClass: AssetClassStorage,
				Confidence: 90,
			}, true
		}
	}
	return FingerprintResult{}, false
}

func isRouterLike(banners map[int]string) (FingerprintResult, bool) {
	for port, b := range banners {
		lower := strings.ToLower(b)
		if strings.Contains(lower, "routeros") || strings.Contains(lower, "mikrotik") || port == 8291 {
			return FingerprintResult{
				OS:         "MikroTik RouterOS",
				OSFamily:   "linux",
				DeviceType: DeviceRouter,
				AssetClass: AssetClassNetwork,
				Confidence: 95,
			}, true
		}
		if strings.Contains(lower, "openwrt") {
			return FingerprintResult{
				OS:         "OpenWrt",
				OSFamily:   "linux",
				DeviceType: DeviceRouter,
				AssetClass: AssetClassNetwork,
				Confidence: 90,
			}, true
		}
		if strings.Contains(lower, "cisco") {
			return FingerprintResult{
				OS:         "Cisco IOS",
				OSFamily:   "unix",
				DeviceType: DeviceRouter,
				AssetClass: AssetClassNetwork,
				Confidence: 85,
			}, true
		}
	}
	return FingerprintResult{}, false
}

func isWindowsLike(banners map[int]string) (FingerprintResult, bool) {
	hasRDP := false
	hasSMB := false
	hasIIS := false
	hasWinSSH := false

	if b, ok := banners[3389]; ok && len(b) > 0 {
		hasRDP = true
	}
	if b, ok := banners[445]; ok && len(b) > 0 {
		hasSMB = true
	}
	if b, ok := banners[22]; ok && strings.Contains(strings.ToLower(b), "openssh_for_windows") {
		hasWinSSH = true
	}
	for port, b := range banners {
		if (port == 80 || port == 443) && strings.Contains(strings.ToLower(b), "iis") {
			hasIIS = true
		}
	}

	// Cross-banner correlations for Windows
	if hasWinSSH || (hasSMB && hasRDP) || (hasSMB && hasIIS) {
		return FingerprintResult{
			OS:         "Windows",
			OSFamily:   "windows",
			DeviceType: DeviceServer,
			AssetClass: AssetClassCompute,
			Confidence: 95,
		}, true
	}

	if hasRDP || hasSMB {
		return FingerprintResult{
			OS:         "Windows",
			OSFamily:   "windows",
			DeviceType: DeviceWorkstation,
			AssetClass: AssetClassCompute,
			Confidence: 75,
		}, true
	}

	return FingerprintResult{}, false
}

func isLinuxLike(banners map[int]string) (FingerprintResult, bool) {
	sshBanner, hasSSH := banners[22]
	if !hasSSH {
		return FingerprintResult{}, false
	}

	lowerSSH := strings.ToLower(sshBanner)
	if strings.Contains(lowerSSH, "ubuntu") {
		return FingerprintResult{
			OS:         "Ubuntu",
			OSFamily:   "linux",
			DeviceType: DeviceServer,
			AssetClass: AssetClassCompute,
			Confidence: 95,
		}, true
	}
	if strings.Contains(lowerSSH, "debian") {
		return FingerprintResult{
			OS:         "Debian",
			OSFamily:   "linux",
			DeviceType: DeviceServer,
			AssetClass: AssetClassCompute,
			Confidence: 95,
		}, true
	}
	if strings.Contains(lowerSSH, "centos") {
		return FingerprintResult{
			OS:         "CentOS",
			OSFamily:   "linux",
			DeviceType: DeviceServer,
			AssetClass: AssetClassCompute,
			Confidence: 95,
		}, true
	}

	return FingerprintResult{}, false
}
