// Package ports implements concurrent TCP port scanning.
//
// Uses connection attempts with internal concurrency control
// via semaphores to prevent file descriptor exhaustion,
// deterministically returning ports that accept active connections.
//
// Main exports:
// - ScanPorts: Concurrently scans a list of ports on a host.
package ports
