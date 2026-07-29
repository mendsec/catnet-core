package fingerprint

// DeviceType defines the category of a network device.
type DeviceType string

const (
	DeviceWorkstation DeviceType = "workstation"
	DeviceServer      DeviceType = "server"
	DeviceRouter      DeviceType = "router"
	DeviceIoT         DeviceType = "iot"
	DeviceMobile      DeviceType = "mobile"
	DeviceUnknown     DeviceType = "unknown"
)

// AssetClass defines broad, high-level asset taxonomy categories.
type AssetClass string

const (
	AssetClassNetwork AssetClass = "network"
	AssetClassCompute AssetClass = "compute"
	AssetClassIoT     AssetClass = "iot"
	AssetClassMobile  AssetClass = "mobile"
	AssetClassStorage AssetClass = "storage"
	AssetClassUnknown AssetClass = "unknown"
)

// BannerGrabConfig controls which active probes are sent during banner grabbing.
type BannerGrabConfig struct {
	// AggressiveSMB sends an SMB negotiate request to port 445.
	// This is an active probe that may trigger IDS/IPS alerts. Default: false.
	AggressiveSMB bool
	// Concurrency limits simultaneous banner grab connections.
	// Default: BannerConcurrency (5).
	Concurrency int
}

const BannerConcurrency = 5

// FingerprintResult holds the detected properties of a host.
type FingerprintResult struct {
	OS         string
	OSFamily   string // "windows", "linux", "macos", "unix", "unknown"
	DeviceType DeviceType
	AssetClass AssetClass
	Vendor     string
	Confidence int // 0-100
}
