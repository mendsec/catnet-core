// Package ports implements concurrent TCP port scanning.
//
// Uses connection attempts with internal concurrency control
// via semaphores to avoid file descriptor exhaustion,
// returning deterministically ports that accept active connections.
//
// Main exports:
// - ScanPorts: Scans a list of ports of a host concurrently.
package ports
