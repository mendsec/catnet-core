package exporter

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/catnet-io/engine/pkg/coreerr"
	"github.com/catnet-io/engine/pkg/results"
	"strconv"
	"strings"
)

// ExportJSON exports scan results to JSON format.
func ExportJSON(report *results.ScanReport) ([]byte, error) {
	out, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to encode JSON: %v", coreerr.ErrExport, err)
	}
	return out, nil
}

// ExportXML exports scan results to XML format.
// Fields per device: IP, Hostname, MAC, Status, OS, DeviceType, AssetClass, Vendor, Open Ports.
// OSFamily is omitted due to redundancy with OS.
func ExportXML(report *results.ScanReport) ([]byte, error) {
	type XMLDevice struct {
		IP         string `xml:"ip"`
		Hostname   string `xml:"hostname"`
		MAC        string `xml:"mac"`
		Status     string `xml:"status"`
		OS         string `xml:"os,omitempty"`
		DeviceType string `xml:"deviceType,omitempty"`
		AssetClass string `xml:"assetClass,omitempty"`
		Vendor     string `xml:"vendor,omitempty"`
	}
	type XMLResults struct {
		XMLName xml.Name    `xml:"results"`
		Devices []XMLDevice `xml:"device"`
	}
	res := XMLResults{}
	for _, d := range report.Devices {
		status := "Dead"
		if d.IsAlive {
			status = "Alive"
		}
		res.Devices = append(res.Devices, XMLDevice{
			IP: d.IP, Hostname: d.Hostname, MAC: d.MAC, Status: status,
			OS: d.OS, DeviceType: d.DeviceType, AssetClass: d.AssetClass, Vendor: d.Vendor,
		})
	}
	out, err := xml.MarshalIndent(res, "", "\t")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to encode XML: %v", coreerr.ErrExport, err)
	}
	return append([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"), out...), nil
}

// sanitizeCSVField sanitizes fields to prevent CSV injection vulnerabilities.
func sanitizeCSVField(field string) string {
	field = strings.ReplaceAll(field, "\n", " ")
	field = strings.ReplaceAll(field, "\r", " ")

	if len(field) == 0 {
		return field
	}

	fc := field[0]
	if fc == '=' || fc == '+' || fc == '-' || fc == '@' || fc == '\t' {
		return "'" + field
	}

	return field
}

// ExportCSV exports scan results to CSV format.
// Fields per device: IP, Hostname, MAC, Status, OS, DeviceType, AssetClass, Vendor, Open Ports.
// OSFamily is omitted due to redundancy with OS.
func ExportCSV(report *results.ScanReport) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{"IP", "Hostname", "MAC", "Status", "OS", "DeviceType", "AssetClass", "Vendor", "Open Ports"}); err != nil {
		return nil, fmt.Errorf("%w: failed to write CSV header: %v", coreerr.ErrExport, err)
	}

	// Reuse string slice to avoid allocation overhead per device
	row := make([]string, 9)
	var portsBuf []byte

	for _, d := range report.Devices {
		row[0] = sanitizeCSVField(d.IP)
		row[1] = sanitizeCSVField(d.Hostname)
		row[2] = sanitizeCSVField(d.MAC)
		if d.IsAlive {
			row[3] = "Alive"
		} else {
			row[3] = "Dead"
		}
		row[4] = sanitizeCSVField(d.OS)
		row[5] = sanitizeCSVField(d.DeviceType)
		row[6] = sanitizeCSVField(d.AssetClass)
		row[7] = sanitizeCSVField(d.Vendor)

		portsBuf = portsBuf[:0]
		for i, p := range d.OpenPorts {
			if i > 0 {
				portsBuf = append(portsBuf, ';')
			}
			portsBuf = strconv.AppendInt(portsBuf, int64(p), 10)
		}
		row[8] = string(portsBuf)

		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("%w: failed to write CSV record: %v", coreerr.ErrExport, err)
		}
	}
	writer.Flush()
	return buf.Bytes(), nil
}
