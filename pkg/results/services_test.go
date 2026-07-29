package results_test

import (
	"testing"

	"github.com/catnet-io/engine/pkg/results"
)

func TestWellKnownServiceName(t *testing.T) {
	tests := []struct {
		port     int
		expected string
	}{
		{20, "FTP-DATA"},
		{21, "FTP"},
		{22, "SSH"},
		{23, "Telnet"},
		{25, "SMTP"},
		{53, "DNS"},
		{67, "DHCP"},
		{68, "DHCP"},
		{69, "TFTP"},
		{80, "HTTP"},
		{110, "POP3"},
		{123, "NTP"},
		{137, "NetBIOS-NS"},
		{138, "NetBIOS-DGM"},
		{139, "NetBIOS-SSN"},
		{143, "IMAP"},
		{161, "SNMP"},
		{389, "LDAP"},
		{443, "HTTPS"},
		{445, "SMB"},
		{465, "SMTPS"},
		{514, "Syslog"},
		{548, "AFP"},
		{587, "SMTP-Submission"},
		{636, "LDAPS"},
		{853, "DNS-over-TLS"},
		{993, "IMAPS"},
		{995, "POP3S"},
		{1433, "MSSQL"},
		{1521, "Oracle DB"},
		{1883, "MQTT"},
		{3306, "MySQL"},
		{3389, "RDP"},
		{5000, "UPnP"},
		{5353, "mDNS"},
		{5432, "PostgreSQL"},
		{5900, "VNC"},
		{6379, "Redis"},
		{8080, "HTTP-Alt"},
		{8443, "HTTPS-Alt"},
		{9000, "SonarQube"},
		{9090, "Prometheus"},
		{9200, "Elasticsearch"},
		{11211, "Memcached"},
		{27017, "MongoDB"},
		{9999, ""},
		{0, ""},
		{-1, ""},
		{65536, ""},
	}

	for _, tt := range tests {
		got := results.WellKnownServiceName(tt.port)
		if got != tt.expected {
			t.Errorf("WellKnownServiceName(%d) = %q; want %q", tt.port, got, tt.expected)
		}
	}
}

func FuzzWellKnownServiceName(f *testing.F) {
	f.Add(22)
	f.Add(80)
	f.Add(443)
	f.Add(-1)
	f.Add(65535)

	f.Fuzz(func(t *testing.T, port int) {
		name := results.WellKnownServiceName(port)
		if port <= 0 || port > 65535 {
			if name != "" {
				t.Errorf("expected empty string for out-of-range port %d, got %q", port, name)
			}
		}
	})
}

func BenchmarkWellKnownServiceName(b *testing.B) {
	testPorts := []int{22, 80, 443, 3389, 8080, 9999}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = results.WellKnownServiceName(testPorts[i%len(testPorts)])
	}
}
