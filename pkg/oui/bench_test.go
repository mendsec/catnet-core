package oui

import "testing"

func BenchmarkLookup(b *testing.B) {
	mac := "B8:27:EB:11:22:33"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Lookup(mac)
	}
}
