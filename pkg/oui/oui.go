package oui

// ouiMap stores the mapping from normalized 6-character uppercase hex OUI prefixes to vendor names.
var ouiMap = map[string]string{
	// Raspberry Pi
	"B827EB": "Raspberry Pi",
	"DCA632": "Raspberry Pi",
	"E45F01": "Raspberry Pi",
	"2CCF67": "Raspberry Pi",

	// Samsung
	"001632": "Samsung",
	"8C71F8": "Samsung",
	"F025B7": "Samsung",
	"5C0E8B": "Samsung",

	// Cisco
	"00000C": "Cisco",
	"000142": "Cisco",
	"0004DD": "Cisco",
	"001B54": "Cisco",

	// Ubiquiti
	"00156D": "Ubiquiti",
	"0418D6": "Ubiquiti",
	"24A43C": "Ubiquiti",
	"B4FBE4": "Ubiquiti",
	"F09FC2": "Ubiquiti",

	// Microsoft
	"0050F2": "Microsoft",
	"281878": "Microsoft",
	"3C8375": "Microsoft",
	"00155D": "Microsoft", // Hyper-V

	// Apple
	"000393": "Apple",
	"000502": "Apple",
	"000A27": "Apple",
	"001124": "Apple",
	"0017F2": "Apple",
	"001C6B": "Apple",
	"001E52": "Apple",
	"001EC2": "Apple",
	"0021E9": "Apple",
	"002241": "Apple",
	"002312": "Apple",
	"002332": "Apple",
	"00236C": "Apple",
	"002436": "Apple",

	// Intel
	"0002B3": "Intel",
	"000347": "Intel",
	"000423": "Intel",
	"000E0C": "Intel",
	"001302": "Intel",

	// Dell
	"00065B": "Dell",
	"000874": "Dell",
	"000BDB": "Dell",
	"001143": "Dell",

	// HP / Hewlett Packard
	"0001E6": "HP",
	"0002A5": "HP",
	"0008C7": "HP",
	"000BCD": "HP",

	// VMware & VirtualBox (Virtualization)
	"000569": "VMware",
	"000C29": "VMware",
	"001C14": "VMware",
	"005056": "VMware",
	"080027": "Oracle/VirtualBox",

	// Network & Storage (Synology, QNAP, TP-Link, Netgear, ASUS, Mikrotik, Arista, Juniper, Fortinet)
	"001132": "Synology",
	"00089B": "QNAP",
	"0014D1": "TP-Link",
	"0019E0": "TP-Link",
	"50C7BF": "TP-Link",
	"00095B": "Netgear",
	"000FB5": "Netgear",
	"000EA6": "ASUS",
	"04D4C4": "ASUS",
	"000C42": "Mikrotik",
	"488E06": "Mikrotik",
	"001C73": "Arista",
	"000585": "Juniper",
	"00090F": "Fortinet",

	// IoT, Camera, Smart Devices (Espressif, Axis, Hikvision, Dahua, Sony, LG, Amazon, Google, Huawei)
	"18FE34": "Espressif",
	"240AC4": "Espressif",
	"30AEA4": "Espressif",
	"600194": "Espressif",
	"00408C": "Axis",
	"001212": "Hikvision",
	"BCAD28": "Hikvision",
	"3C15C2": "Dahua",
	"00014A": "Sony",
	"000B97": "LG",
	"44650D": "Amazon",
	"FCA667": "Amazon",
	"001A11": "Google",
	"D8EB46": "Google",
	"000B45": "Huawei",
	"001E10": "Huawei",
}

// Lookup returns the vendor name associated with the MAC address prefix.
// Returns an empty string if the MAC address is invalid or the OUI is unrecognized.
func Lookup(mac string) string {
	vendor, _ := LookupWithPrefix(mac)
	return vendor
}

// LookupWithPrefix returns the vendor name and a boolean indicating whether the OUI prefix was found in the offline database.
func LookupWithPrefix(mac string) (string, bool) {
	// Skip leading whitespace without extra allocations.
	start := 0
	for start < len(mac) {
		switch mac[start] {
		case ' ', '\t', '\r', '\n':
			start++
		default:
			goto parse
		}
	}
	return "", false

parse:
	trimmed := mac[start:]

	// Extract hex characters for the 24-bit OUI (6 hex digits).
	var ouiBuf [6]byte
	count := 0

	for i := 0; i < len(trimmed) && count < 6; i++ {
		c := trimmed[i]
		if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') {
			ouiBuf[count] = c
			count++
		} else if c >= 'a' && c <= 'z' {
			ouiBuf[count] = c - ('a' - 'A')
			count++
		} else if c == ':' || c == '-' || c == '.' {
			continue
		} else {
			// Invalid character in MAC address
			return "", false
		}
	}

	if count < 6 {
		return "", false
	}

	vendor, ok := ouiMap[string(ouiBuf[:])]
	return vendor, ok
}
