package fingerprint

import (
	"github.com/catnet-io/engine/pkg/oui"
)

// VendorFromMAC returns the vendor name from the MAC address using the built-in IEEE OUI database in pkg/oui.
func VendorFromMAC(mac string) string {
	return oui.Lookup(mac)
}
