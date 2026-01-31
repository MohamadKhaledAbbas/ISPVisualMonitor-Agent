package mikrotik

import (
	"context"
	"math/rand"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/collector/mikrotik/api"
)

// NATConnection represents a NAT connection entry.
type NATConnection struct {
	Protocol    string `json:"protocol"`
	SrcAddress  string `json:"src_address"`
	SrcPort     int64  `json:"src_port,omitempty"`
	DstAddress  string `json:"dst_address"`
	DstPort     int64  `json:"dst_port,omitempty"`
	ReplyAddr   string `json:"reply_addr,omitempty"`
	ReplyPort   int64  `json:"reply_port,omitempty"`
	TCPState    string `json:"tcp_state,omitempty"`
	Timeout     int64  `json:"timeout_seconds,omitempty"`
	Assured     bool   `json:"assured,omitempty"`
	Confirmed   bool   `json:"confirmed,omitempty"`
	Dying       bool   `json:"dying,omitempty"`
	FastTracked bool   `json:"fast_tracked,omitempty"`
	GREKey      string `json:"gre_key,omitempty"`
	GREVersion  string `json:"gre_version,omitempty"`
	ICMPCode    int64  `json:"icmp_code,omitempty"`
	ICMPType    int64  `json:"icmp_type,omitempty"`
	ICMPId      int64  `json:"icmp_id,omitempty"`
	RxBytes     int64  `json:"rx_bytes,omitempty"`
	TxBytes     int64  `json:"tx_bytes,omitempty"`
	RxPackets   int64  `json:"rx_packets,omitempty"`
	TxPackets   int64  `json:"tx_packets,omitempty"`
}

// NATStats contains NAT statistics.
type NATStats struct {
	TotalConnections  int   `json:"total_connections"`
	SampledCount      int   `json:"sampled_count,omitempty"`
	TCPConnections    int   `json:"tcp_connections"`
	UDPConnections    int   `json:"udp_connections"`
	ICMPConnections   int   `json:"icmp_connections"`
	OtherConnections  int   `json:"other_connections"`
	MaxEntries        int64 `json:"max_entries,omitempty"`
	TotalTCPEntries   int64 `json:"total_tcp_entries,omitempty"`
	TotalUDPEntries   int64 `json:"total_udp_entries,omitempty"`
	TotalICMPEntries  int64 `json:"total_icmp_entries,omitempty"`
}

// collectNAT collects NAT/connection tracking information from the router.
func (c *Collector) collectNAT(ctx context.Context, client *api.Client) ([]NATConnection, *NATStats, error) {
	stats := &NATStats{}

	// Get connection tracking stats first
	ctStats, err := client.RunOne(ctx, "/ip/firewall/connection/tracking/print", nil)
	if err == nil && ctStats != nil {
		stats.MaxEntries = ParseInt64(ctStats["max-entries"])
		stats.TotalTCPEntries = ParseInt64(ctStats["total-tcp-entries"])
		stats.TotalUDPEntries = ParseInt64(ctStats["total-udp-entries"])
		stats.TotalICMPEntries = ParseInt64(ctStats["total-icmp-entries"])
	}

	// Get active connections
	connections, err := client.Run(ctx, "/ip/firewall/connection/print", nil)
	if err != nil {
		return nil, stats, err
	}

	stats.TotalConnections = len(connections)

	var result []NATConnection
	maxConns := c.config.NAT.MaxConnections
	if maxConns <= 0 {
		maxConns = 10000
	}

	for i, conn := range connections {
		// Apply sampling if enabled
		if c.config.NAT.SamplingEnabled && c.config.NAT.SampleRate < 1.0 {
			if rand.Float64() > c.config.NAT.SampleRate {
				continue
			}
		}

		// Apply max connections limit
		if len(result) >= maxConns {
			break
		}

		natConn := parseNATConnection(conn)

		// Count by protocol
		switch natConn.Protocol {
		case "tcp":
			stats.TCPConnections++
		case "udp":
			stats.UDPConnections++
		case "icmp":
			stats.ICMPConnections++
		default:
			stats.OtherConnections++
		}

		result = append(result, natConn)

		// Safety check for very large connection tables
		if i > 100000 {
			break
		}
	}

	stats.SampledCount = len(result)

	return result, stats, nil
}

func parseNATConnection(conn map[string]string) NATConnection {
	natConn := NATConnection{
		Protocol:    conn["protocol"],
		SrcAddress:  conn["src-address"],
		DstAddress:  conn["dst-address"],
		ReplyAddr:   conn["reply-src-address"],
		TCPState:    conn["tcp-state"],
		Timeout:     parseTimeout(conn["timeout"]),
		Assured:     conn["assured"] == "true",
		Confirmed:   conn["confirmed"] == "true",
		Dying:       conn["dying"] == "true",
		FastTracked: conn["fasttrack"] == "true",
		GREKey:      conn["gre-key"],
		GREVersion:  conn["gre-version"],
	}

	// Parse ports from address:port format
	natConn.SrcAddress, natConn.SrcPort = parseAddressPort(conn["src-address"])
	natConn.DstAddress, natConn.DstPort = parseAddressPort(conn["dst-address"])
	natConn.ReplyAddr, natConn.ReplyPort = parseAddressPort(conn["reply-src-address"])

	// ICMP specific
	natConn.ICMPCode = ParseInt64(conn["icmp-code"])
	natConn.ICMPType = ParseInt64(conn["icmp-type"])
	natConn.ICMPId = ParseInt64(conn["icmp-id"])

	// Traffic counters
	if orig := conn["orig-bytes"]; orig != "" {
		natConn.TxBytes = ParseInt64(orig)
	}
	if reply := conn["repl-bytes"]; reply != "" {
		natConn.RxBytes = ParseInt64(reply)
	}
	if orig := conn["orig-packets"]; orig != "" {
		natConn.TxPackets = ParseInt64(orig)
	}
	if reply := conn["repl-packets"]; reply != "" {
		natConn.RxPackets = ParseInt64(reply)
	}

	return natConn
}

// parseAddressPort parses "address:port" format.
func parseAddressPort(s string) (string, int64) {
	// Find the last colon (for IPv6 compatibility)
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == ':' {
			addr := s[:i]
			port := ParseInt64(s[i+1:])
			return addr, port
		}
	}
	return s, 0
}

// parseTimeout parses RouterOS timeout string to seconds.
func parseTimeout(s string) int64 {
	return ParseUptime(s)
}
