// Package targets provides functionality for network target parsing.
//
// The package is responsible for interpreting formats such as CIDR, IP ranges
// (dash ranges), individual IP addresses, and local interface auto-detection ("auto"),
// expanding them into valid lists of IP addresses ready for engine scanning.
//
// Primary exports:
// - ParseRange: Parses and converts a target string (including "auto") into a list of IPs.
// - DetectLocalRange: Detects active non-loopback local network interface subnets and returns their IPs.
package targets

