package discovery

import (
	"context"
	"net"
	"strings"

	"github.com/catnet-io/engine/internal/netutil"
)

// Ping performs an ICMP ping on the machine with a timeout in milliseconds.
func Ping(ctx context.Context, ip string, timeoutMs int) bool {
	if err := netutil.ValidateIPv4(ip); err != nil {
		return false
	}
	return osPing(ctx, ip, timeoutMs)
}

// ReverseDNS resolves the hostname of the given IP address.
func ReverseDNS(ctx context.Context, ip string) string {
	if err := netutil.ValidateIPv4(ip); err != nil {
		return ""
	}
	names, err := net.DefaultResolver.LookupAddr(ctx, ip)
	if err == nil && len(names) > 0 {
		return strings.TrimSuffix(names[0], ".")
	}
	return ""
}

// GetMAC attempts to obtain the MAC address of the target machine.
func GetMAC(ctx context.Context, ip string) string {
	if err := netutil.ValidateIPv4(ip); err != nil {
		return ""
	}
	return osGetMAC(ctx, ip)
}
