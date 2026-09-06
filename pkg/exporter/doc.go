// Package exporter handles the formatting and serialization of reports.
//
// Supports the conversion of ScanReport into standardized formats (JSON, XML, CSV)
// ensuring that exports avoid security issues like CSV Injection.
//
// Main exports:
// - ExportJSON: Exports reports as indented JSON.
// - ExportXML: Exports reports as valid XML.
// - ExportCSV: Exports reports to CSV, handling formula injection.
package exporter
