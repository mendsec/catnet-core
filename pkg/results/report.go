package results

import "time"

// ScanReport encapsulates the complete result of a scan.
type ScanReport struct {
	SchemaVersion string       `json:"schemaVersion"`
	StartTime     time.Time    `json:"startTime"`
	EndTime       time.Time    `json:"endTime"`
	Total         int          `json:"total"`
	Alive         int          `json:"alive"`
	Devices       []DeviceInfo `json:"devices"`
}

// NewScanReport creates a new scan report.
func NewScanReport() *ScanReport {
	return &ScanReport{
		SchemaVersion: "2.0.0",
		StartTime:     time.Now(),
		Devices:       []DeviceInfo{},
	}
}
