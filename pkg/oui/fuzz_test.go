package oui

import "testing"

func FuzzLookup(f *testing.F) {
	f.Add("B8:27:EB:00:00:00")
	f.Add("dc-a6-32-11-22-33")
	f.Add("0050.56aa.bbcc")
	f.Add("invalid_mac_string")
	f.Add("")
	f.Add("00:00:00")

	f.Fuzz(func(t *testing.T, mac string) {
		vendor, ok := LookupWithPrefix(mac)
		lookupVendor := Lookup(mac)
		if lookupVendor != vendor {
			t.Errorf("Lookup(%q) != LookupWithPrefix(%q) vendor: got %q vs %q", mac, mac, lookupVendor, vendor)
		}
		if ok && vendor == "" {
			t.Errorf("LookupWithPrefix(%q) returned ok=true but vendor is empty", mac)
		}
	})
}
