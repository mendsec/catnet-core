package results

type HostResult struct {
	IP        string `json:"ip"`
	Alive     bool   `json:"alive"`
	Hostname  string `json:"hostname"`
	MAC       string `json:"mac"`
	OpenPorts []int  `json:"open_ports"`
}

func (h HostResult) ToDeviceInfo() DeviceInfo {
	return DeviceInfo{
		IP:        h.IP,
		IsAlive:   h.Alive,
		Hostname:  h.Hostname,
		MAC:       h.MAC,
		OpenPorts: h.OpenPorts,
	}
}

type DeviceInfo struct {
	IP         string `json:"ip"`
	IsAlive    bool   `json:"isAlive"`
	Hostname   string `json:"hostname"`
	MAC        string `json:"mac"`
	OS         string `json:"os,omitempty"`
	OSFamily   string `json:"osFamily,omitempty"`
	DeviceType string `json:"deviceType,omitempty"`
	AssetClass string `json:"assetClass,omitempty"`
	Vendor     string `json:"vendor,omitempty"`
	OpenPorts  []int  `json:"openPorts"`
}

func (d DeviceInfo) PortCount() int {
	return len(d.OpenPorts)
}
