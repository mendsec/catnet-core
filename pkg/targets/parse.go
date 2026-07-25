package targets

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"github.com/catnet-io/engine/pkg/coreerr"
)

// ParseRange parses an input string and returns a slice of matching IP addresses.
// It supports single IPs, CIDR ranges, dash ranges, and "auto" target auto-detection.
func ParseRange(input string) ([]string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("%w: empty input", coreerr.ErrInvalidInput)
	}
	if strings.EqualFold(input, "auto") {
		return DetectLocalRange()
	}
	if strings.Contains(input, "/") {
		return parseCIDR(input)
	}
	if strings.Contains(input, "-") {
		return parseDashRange(input)
	}
	parsed := net.ParseIP(input)
	if parsed == nil {
		return nil, fmt.Errorf("%w: invalid IP format", coreerr.ErrInvalidInput)
	}
	return []string{parsed.String()}, nil
}

// DetectLocalRange auto-detects active non-loopback local IPv4 network interface subnets
// and returns the expanded list of matching IP addresses.
func DetectLocalRange() ([]string, error) {
	return detectLocalRangeCustom(net.Interfaces, func(iface net.Interface) ([]net.Addr, error) {
		return iface.Addrs()
	})
}

func detectLocalRangeCustom(
	getIfaces func() ([]net.Interface, error),
	getAddrs func(iface net.Interface) ([]net.Addr, error),
) ([]string, error) {
	ifaces, err := getIfaces()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to list network interfaces: %v", coreerr.ErrInvalidInput, err)
	}

	seenIPs := make(map[string]struct{})
	var result []string

	for _, iface := range ifaces {
		// Filter out down or loopback interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := getAddrs(iface)
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet == nil || ipNet.IP == nil {
				continue
			}

			ip4 := ipNet.IP.To4()
			if ip4 == nil || ip4.IsLoopback() || ip4.IsLinkLocalUnicast() || ip4.IsUnspecified() {
				continue
			}

			ips, err := parseCIDR(ipNet.String())
			if err != nil {
				continue
			}

			for _, ipStr := range ips {
				if _, seen := seenIPs[ipStr]; !seen {
					seenIPs[ipStr] = struct{}{}
					result = append(result, ipStr)
				}
			}
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("%w: no active local network interfaces found", coreerr.ErrInvalidInput)
	}

	return result, nil
}


func parseCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	// Bolt Optimization: Pre-allocate slice to prevent dynamic resizing overhead.
	ones, bits := ipnet.Mask.Size()
	if bits-ones > 16 {
		return nil, fmt.Errorf("range too large (max 65536)")
	}
	capacity := 1 << (bits - ones)
	ips := make([]string, 0, capacity)
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); {
		ips = append(ips, ip.String())
		if inc(ip) {
			break
		}
	}
	if len(ips) > 2 { // Skip network and broadcast addresses
		return ips[1 : len(ips)-1], nil
	}
	return ips, nil
}

func parseDashRange(dashStr string) ([]string, error) {
	parts := strings.Split(dashStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w: invalid range format, expected start-end", coreerr.ErrInvalidInput)
	}
	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])
	if !strings.Contains(endStr, ".") {
		lastDot := strings.LastIndex(startStr, ".")
		if lastDot != -1 {
			endStr = startStr[:lastDot+1] + endStr
		}
	}
	startIP := net.ParseIP(startStr).To4()
	endIP := net.ParseIP(endStr).To4()
	if startIP == nil || endIP == nil {
		return nil, fmt.Errorf("%w: invalid IP in range", coreerr.ErrInvalidInput)
	}
	start := binary.BigEndian.Uint32(startIP)
	end := binary.BigEndian.Uint32(endIP)
	if start > end {
		return nil, fmt.Errorf("%w: start IP is greater than end IP", coreerr.ErrInvalidInput)
	}
	if end-start > 65536 {
		return nil, fmt.Errorf("%w: range too large (max 65536)", coreerr.ErrInvalidInput)
	}
	// Bolt Optimization: Pre-allocate slice and reuse IP buffer to reduce memory allocations.
	capacity := end - start + 1
	ips := make([]string, 0, capacity)
	ip := make(net.IP, 4)
	for i := start; i <= end; i++ {
		binary.BigEndian.PutUint32(ip, i)
		ips = append(ips, ip.String())
	}
	return ips, nil
}

func inc(ip net.IP) bool {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			return false
		}
	}
	return true
}
