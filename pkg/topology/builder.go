package topology

import (
	"github.com/catnet-io/engine/pkg/results"
)

// BuildGraph creates an adjacency graph from scan report.
func BuildGraph(report *results.ScanReport) *TopologyGraph {
	if report == nil {
		return &TopologyGraph{
			Nodes: []TopologyNode{},
			Edges: []TopologyEdge{},
		}
	}

	graph := &TopologyGraph{
		Nodes: make([]TopologyNode, 0, len(report.Devices)),
		Edges: make([]TopologyEdge, 0),
	}

	gateway := DetectGateway(report.Devices)
	graph.Gateway = gateway

	// 2. Create nodes
	for _, dev := range report.Devices {
		var role NodeRole
		if dev.IP == gateway && gateway != "" {
			role = RoleGateway
		} else if !dev.IsAlive {
			role = RoleUnknown
		} else if dev.DeviceType == "server" {
			role = RoleServer
		} else {
			role = RoleHost
		}

		label := dev.Hostname
		if label == "" {
			label = dev.IP
		}

		node := TopologyNode{
			ID:         dev.IP,
			Label:      label,
			Role:       role,
			DeviceType: dev.DeviceType,
			IsAlive:    dev.IsAlive,
			OpenPorts:  make([]int, len(dev.OpenPorts)),
		}
		copy(node.OpenPorts, dev.OpenPorts)
		graph.Nodes = append(graph.Nodes, node)
	}

	// 3. Create Edges
	// gateway -> host for all alive hosts (weight 1.0)
	// host -> host when sharing a service in the same /24 subnet (weight 0.3)

	// O(n) mapping of subnets and open ports
	subnetMap := make(map[string]map[int][]string) // subnet -> port -> []ips

	for _, dev := range report.Devices {
		if !dev.IsAlive || dev.IP == gateway {
			continue
		}

		if gateway != "" {
			graph.Edges = append(graph.Edges, TopologyEdge{
				Source: gateway,
				Target: dev.IP,
				Weight: 1.0,
			})
		}

		// Find /24 subnet
		// ⚡ Bolt Optimization: Use zero-allocation counting and slicing
		// Avoids massive allocation overhead and ensures exactly 3 dots (valid IPv4)
		dots := 0
		thirdDotIdx := -1
		for i := 0; i < len(dev.IP); i++ {
			if dev.IP[i] == '.' {
				dots++
				if dots == 3 {
					thirdDotIdx = i
				}
			}
		}

		if dots == 3 && thirdDotIdx != -1 {
			subnet := dev.IP[:thirdDotIdx]
			if subnetMap[subnet] == nil {
				subnetMap[subnet] = make(map[int][]string)
			}
			for _, port := range dev.OpenPorts {
				subnetMap[subnet][port] = append(subnetMap[subnet][port], dev.IP)
			}
		}
	}

	// host -> host edges
	const maxEdgesPerSubnet = 200
	type edgeKey struct {
		src, dst string
	}
	addedHostEdges := make(map[edgeKey]struct{})
	for _, portsMap := range subnetMap {
		edgesInSubnet := 0
		for _, ipList := range portsMap {
			if edgesInSubnet >= maxEdgesPerSubnet {
				break
			}
			if len(ipList) > 1 {
				for i := 0; i < len(ipList); i++ {
					if edgesInSubnet >= maxEdgesPerSubnet {
						break
					}
					for j := i + 1; j < len(ipList); j++ {
						if edgesInSubnet >= maxEdgesPerSubnet {
							break
						}
						src := ipList[i]
						dst := ipList[j]
						if src == dst {
							continue
						}
						// ensure src < dst to avoid duplicate bidirectional edges
						if src > dst {
							src, dst = dst, src
						}
						// ⚡ Bolt Optimization: Use zero-allocation struct key instead of string concatenation.
						// This prevents massive memory allocation and GC overhead in dense O(N^2) graphs.
						key := edgeKey{src, dst}
						if _, exists := addedHostEdges[key]; !exists {
							addedHostEdges[key] = struct{}{}
							graph.Edges = append(graph.Edges, TopologyEdge{
								Source: src,
								Target: dst,
								Weight: 0.3,
							})
							edgesInSubnet++
						}
					}
				}
			}
		}
	}

	return graph
}
