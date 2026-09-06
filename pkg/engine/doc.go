// Package engine orchestrates concurrent network scans.
//
// The package is the main entry point for catnet-core, providing the
// StartScan function to manage goroutine pools, timeouts,
// and the emission of asynchronous events during the scan.
//
// Main exports:
// - StartScan: Starts a network scan.
// - ScanConfig: Configuration such as concurrency limits and timeouts.
// - EventCallback: Type for receiving progress and result events.
package engine
