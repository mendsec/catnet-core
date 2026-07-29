package results

import "github.com/catnet-io/engine/pkg/ports"

// WellKnownServiceName returns the well-known service name associated with a given TCP port number.
// If the port is not recognized, it returns an empty string.
func WellKnownServiceName(port int) string {
	return ports.WellKnownServiceName(port)
}
