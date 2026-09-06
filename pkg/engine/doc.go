// Package engine orchestrates the concurrent execution of network scans.
//
// The package is the main entry point for catnet-core, providing the
// StartScan function to manage goroutine pools, execution timeouts
// and the emission of asynchronous events during the scan.
//
// Main exports:
// - StartScan: Initiates a network scan.
// - ScanConfig: Settings such as concurrency limits and timeouts.
// - EventCallback: Type for receiving progress and result events.
package engine
