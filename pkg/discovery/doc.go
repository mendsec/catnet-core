// Package discovery implements the detection of host attributes on the network.
//
// Offers cross-platform abstractions for liveness detection via ICMP (Ping),
// reverse DNS resolution and MAC address retrieval through the local ARP table.
// Specific implementations for Windows and POSIX are available internally.
//
// Main exports:
// - Ping: Sends an ICMP Echo request to check if a host is active.
// - ReverseDNS: Obtains the hostname associated with an IP address.
// - GetMAC: Obtains the MAC address corresponding to an IP.
package discovery
