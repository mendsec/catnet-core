package oui

import "testing"

func TestLookup(t *testing.T) {
	tests := []struct {
		name       string
		mac        string
		wantVendor string
		wantFound  bool
	}{
		{
			name:       "Raspberry Pi standard colon format",
			mac:        "B8:27:EB:00:00:00",
			wantVendor: "Raspberry Pi",
			wantFound:  true,
		},
		{
			name:       "Raspberry Pi hyphen format lowercase",
			mac:        "dc-a6-32-11-22-33",
			wantVendor: "Raspberry Pi",
			wantFound:  true,
		},
		{
			name:       "Raspberry Pi dot format",
			mac:        "e45f.0111.2233",
			wantVendor: "Raspberry Pi",
			wantFound:  true,
		},
		{
			name:       "VMware colon format",
			mac:        "00:50:56:AA:BB:CC",
			wantVendor: "VMware",
			wantFound:  true,
		},
		{
			name:       "Oracle/VirtualBox colon format",
			mac:        "08:00:27:12:34:56",
			wantVendor: "Oracle/VirtualBox",
			wantFound:  true,
		},
		{
			name:       "Apple colon format with leading whitespace",
			mac:        "  00:1E:C2:55:66:77  ",
			wantVendor: "Apple",
			wantFound:  true,
		},
		{
			name:       "Espressif ESP32",
			mac:        "18:FE:34:00:11:22",
			wantVendor: "Espressif",
			wantFound:  true,
		},
		{
			name:       "Unknown valid OUI",
			mac:        "FE:FF:FF:00:00:00",
			wantVendor: "",
			wantFound:  false,
		},
		{
			name:       "Malformed short MAC",
			mac:        "B8:27",
			wantVendor: "",
			wantFound:  false,
		},
		{
			name:       "Invalid characters in MAC",
			mac:        "B8:27:G9:00:00:00",
			wantVendor: "",
			wantFound:  false,
		},
		{
			name:       "Empty MAC string",
			mac:        "",
			wantVendor: "",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVendor := Lookup(tt.mac)
			if gotVendor != tt.wantVendor {
				t.Errorf("Lookup(%q) = %q; want %q", tt.mac, gotVendor, tt.wantVendor)
			}

			vendor, ok := LookupWithPrefix(tt.mac)
			if vendor != tt.wantVendor || ok != tt.wantFound {
				t.Errorf("LookupWithPrefix(%q) = (%q, %v); want (%q, %v)", tt.mac, vendor, ok, tt.wantVendor, tt.wantFound)
			}
		})
	}
}
