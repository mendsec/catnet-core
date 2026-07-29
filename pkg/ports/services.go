package ports

import "github.com/catnet-io/engine/pkg/results"

// WellKnownServiceName returns the common service name associated with a given TCP port number.
// It delegates to results.WellKnownServiceName.
func WellKnownServiceName(port int) string {
	return results.WellKnownServiceName(port)
}
