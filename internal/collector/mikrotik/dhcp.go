package mikrotik

import (
	"context"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/collector/mikrotik/api"
)

// DHCPLease represents a DHCP lease entry.
type DHCPLease struct {
	ID           string    `json:"id"`
	Address      string    `json:"address"`
	MACAddress   string    `json:"mac_address"`
	Hostname     string    `json:"hostname,omitempty"`
	Comment      string    `json:"comment,omitempty"`
	ServerName   string    `json:"server_name,omitempty"`
	Status       string    `json:"status,omitempty"`   // bound, waiting, offered
	ExpiresAfter int64     `json:"expires_after,omitempty"` // Seconds until expiry
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	LastSeen     time.Time `json:"last_seen,omitempty"`
	ActiveServer string    `json:"active_server,omitempty"`
	Blocked      bool      `json:"blocked,omitempty"`
	Disabled     bool      `json:"disabled,omitempty"`
	Dynamic      bool      `json:"dynamic,omitempty"`
	RateLimit    string    `json:"rate_limit,omitempty"`
	AddressLists string    `json:"address_lists,omitempty"`
}

// DHCPPoolStats contains DHCP pool statistics.
type DHCPPoolStats struct {
	Name           string  `json:"name"`
	Ranges         string  `json:"ranges,omitempty"`
	TotalAddresses int64   `json:"total_addresses"`
	UsedAddresses  int64   `json:"used_addresses"`
	FreeAddresses  int64   `json:"free_addresses"`
	Utilization    float64 `json:"utilization_percent"`
}

// DHCPServerStats contains DHCP server statistics.
type DHCPServerStats struct {
	Name          string `json:"name"`
	Interface     string `json:"interface"`
	AddressPool   string `json:"address_pool"`
	LeaseTime     string `json:"lease_time,omitempty"`
	Disabled      bool   `json:"disabled"`
	Authoritative bool   `json:"authoritative,omitempty"`
	ActiveLeases  int    `json:"active_leases"`
	TotalLeases   int    `json:"total_leases"`
}

// collectDHCP collects DHCP lease information from the router.
func (c *Collector) collectDHCP(ctx context.Context, client *api.Client) ([]DHCPLease, []DHCPPoolStats, []DHCPServerStats, error) {
	// Get DHCP leases
	leases, err := client.Run(ctx, "/ip/dhcp-server/lease/print", nil)
	if err != nil {
		return nil, nil, nil, err
	}

	var leaseList []DHCPLease
	serverLeaseCounts := make(map[string]int)
	serverActiveCounts := make(map[string]int)

	now := time.Now()

	for _, l := range leases {
		lease := DHCPLease{
			ID:           l[".id"],
			Address:      l["address"],
			MACAddress:   l["mac-address"],
			Hostname:     l["host-name"],
			Comment:      l["comment"],
			ServerName:   l["server"],
			Status:       l["status"],
			ExpiresAfter: ParseUptime(l["expires-after"]),
			ActiveServer: l["active-server"],
			Blocked:      l["blocked"] == "true",
			Disabled:     l["disabled"] == "true",
			Dynamic:      l["dynamic"] == "true",
			RateLimit:    l["rate-limit"],
			AddressLists: l["address-lists"],
		}

		// Calculate expiry time
		if lease.ExpiresAfter > 0 {
			lease.ExpiresAt = now.Add(time.Duration(lease.ExpiresAfter) * time.Second)
		}

		// Parse last seen
		if lastSeen := l["last-seen"]; lastSeen != "" {
			lease.LastSeen = ParseTimestamp(lastSeen)
		}

		leaseList = append(leaseList, lease)

		// Count leases per server
		if lease.ServerName != "" {
			serverLeaseCounts[lease.ServerName]++
			if lease.Status == "bound" {
				serverActiveCounts[lease.ServerName]++
			}
		}
	}

	// Get DHCP pools
	pools, err := client.Run(ctx, "/ip/pool/print", nil)
	if err != nil {
		// Pools might not exist, continue
		pools = nil
	}

	var poolStats []DHCPPoolStats
	poolUsage := make(map[string]int64)

	// Count used addresses per pool
	poolUsedAddrs, _ := client.Run(ctx, "/ip/pool/used/print", nil)
	for _, used := range poolUsedAddrs {
		poolName := used["pool"]
		poolUsage[poolName]++
	}

	for _, p := range pools {
		name := p["name"]
		ranges := p["ranges"]

		stats := DHCPPoolStats{
			Name:           name,
			Ranges:         ranges,
			TotalAddresses: countPoolAddresses(ranges),
			UsedAddresses:  poolUsage[name],
		}
		stats.FreeAddresses = stats.TotalAddresses - stats.UsedAddresses
		if stats.TotalAddresses > 0 {
			stats.Utilization = float64(stats.UsedAddresses) / float64(stats.TotalAddresses) * 100
		}

		poolStats = append(poolStats, stats)
	}

	// Get DHCP servers
	servers, err := client.Run(ctx, "/ip/dhcp-server/print", nil)
	if err != nil {
		// DHCP server might not be configured
		servers = nil
	}

	var serverStats []DHCPServerStats

	for _, s := range servers {
		name := s["name"]
		stats := DHCPServerStats{
			Name:          name,
			Interface:     s["interface"],
			AddressPool:   s["address-pool"],
			LeaseTime:     s["lease-time"],
			Disabled:      s["disabled"] == "true",
			Authoritative: s["authoritative"] == "yes" || s["authoritative"] == "true",
			TotalLeases:   serverLeaseCounts[name],
			ActiveLeases:  serverActiveCounts[name],
		}

		serverStats = append(serverStats, stats)
	}

	return leaseList, poolStats, serverStats, nil
}

// countPoolAddresses counts the total number of addresses in a pool range.
// Range format: "192.168.1.10-192.168.1.100" or "192.168.1.10-192.168.1.100,192.168.2.10-192.168.2.50"
func countPoolAddresses(ranges string) int64 {
	if ranges == "" {
		return 0
	}

	var total int64

	// Simple implementation - count based on last octet difference
	// A more complete implementation would parse full IP addresses
	for _, r := range splitRanges(ranges) {
		start, end := parseIPRange(r)
		if end >= start {
			total += int64(end - start + 1)
		}
	}

	return total
}

func splitRanges(s string) []string {
	var ranges []string
	var current string
	
	for _, c := range s {
		if c == ',' {
			if current != "" {
				ranges = append(ranges, current)
			}
			current = ""
		} else if c != ' ' {
			current += string(c)
		}
	}
	if current != "" {
		ranges = append(ranges, current)
	}
	
	return ranges
}

func parseIPRange(r string) (int, int) {
	// Find the dash separator
	var start, end string
	dashFound := false
	
	for i, c := range r {
		if c == '-' && !dashFound {
			start = r[:i]
			end = r[i+1:]
			dashFound = true
			break
		}
	}
	
	if !dashFound {
		return 0, 0
	}

	startNum := parseLastOctet(start)
	endNum := parseLastOctet(end)
	
	return startNum, endNum
}

func parseLastOctet(ip string) int {
	// Find the last dot and parse the number after it
	lastDot := -1
	for i := len(ip) - 1; i >= 0; i-- {
		if ip[i] == '.' {
			lastDot = i
			break
		}
	}
	
	if lastDot == -1 {
		return 0
	}
	
	return int(ParseInt64(ip[lastDot+1:]))
}
