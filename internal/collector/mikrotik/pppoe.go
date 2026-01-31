package mikrotik

import (
	"context"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/collector/mikrotik/api"
)

// PPPoESession represents a PPPoE session.
type PPPoESession struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Service        string `json:"service,omitempty"`
	Username       string `json:"username"`
	CallerID       string `json:"caller_id,omitempty"`       // MAC address
	Address        string `json:"address,omitempty"`         // Assigned IP
	Uptime         int64  `json:"uptime_seconds,omitempty"`
	RxBytes        int64  `json:"rx_bytes,omitempty"`
	TxBytes        int64  `json:"tx_bytes,omitempty"`
	RxPackets      int64  `json:"rx_packets,omitempty"`
	TxPackets      int64  `json:"tx_packets,omitempty"`
	RateLimit      string `json:"rate_limit,omitempty"`
	SessionID      string `json:"session_id,omitempty"`
	MTU            int64  `json:"mtu,omitempty"`
	MRU            int64  `json:"mru,omitempty"`
	Encoding       string `json:"encoding,omitempty"`
	LimitBytesIn   int64  `json:"limit_bytes_in,omitempty"`
	LimitBytesOut  int64  `json:"limit_bytes_out,omitempty"`
}

// PPPoEServerStats contains PPPoE server statistics.
type PPPoEServerStats struct {
	ServerName     string `json:"server_name"`
	Interface      string `json:"interface"`
	ActiveSessions int    `json:"active_sessions"`
	TotalSessions  int    `json:"total_sessions"`
}

// collectPPPoE collects PPPoE session information from the router.
func (c *Collector) collectPPPoE(ctx context.Context, client *api.Client) ([]PPPoESession, []PPPoEServerStats, error) {
	// Get active PPPoE sessions
	sessions, err := client.Run(ctx, "/ppp/active/print", nil)
	if err != nil {
		return nil, nil, err
	}

	var pppoeList []PPPoESession

	for _, s := range sessions {
		session := PPPoESession{
			ID:            s[".id"],
			Name:          s["name"],
			Service:       s["service"],
			Username:      s["name"], // PPPoE username is typically the connection name
			CallerID:      s["caller-id"],
			Address:       s["address"],
			Uptime:        ParseUptime(s["uptime"]),
			RxBytes:       ParseInt64(s["bytes"]),   // Combined or use limit-bytes-in/out
			TxBytes:       ParseInt64(s["bytes"]),   // Will be separated below
			RxPackets:     ParseInt64(s["packets"]),
			TxPackets:     ParseInt64(s["packets"]),
			RateLimit:     s["rate-limit"],
			SessionID:     s["session-id"],
			MTU:           ParseInt64(s["mtu"]),
			MRU:           ParseInt64(s["mru"]),
			Encoding:      s["encoding"],
			LimitBytesIn:  ParseInt64(s["limit-bytes-in"]),
			LimitBytesOut: ParseInt64(s["limit-bytes-out"]),
		}

		// Parse bytes if in "rx,tx" format
		if bytesStr := s["bytes"]; bytesStr != "" {
			session.RxBytes, session.TxBytes = parseBytePair(bytesStr)
		}
		if packetsStr := s["packets"]; packetsStr != "" {
			session.RxPackets, session.TxPackets = parseBytePair(packetsStr)
		}

		pppoeList = append(pppoeList, session)
	}

	// Get PPPoE server statistics
	servers, err := client.Run(ctx, "/interface/pppoe-server/server/print", nil)
	if err != nil {
		// PPPoE server might not be configured, continue without error
		return pppoeList, nil, nil
	}

	var serverStats []PPPoEServerStats

	for _, srv := range servers {
		stats := PPPoEServerStats{
			ServerName: srv["service-name"],
			Interface:  srv["interface"],
		}

		// Count active sessions for this server
		for _, s := range pppoeList {
			if s.Service == stats.ServerName {
				stats.ActiveSessions++
			}
		}

		serverStats = append(serverStats, stats)
	}

	return pppoeList, serverStats, nil
}

// parseBytePair parses "rx,tx" format byte strings.
func parseBytePair(s string) (int64, int64) {
	var rx, tx int64
	
	// RouterOS returns bytes as "rx,tx" or just a single value
	for i, c := range s {
		if c == ',' {
			rx = ParseInt64(s[:i])
			tx = ParseInt64(s[i+1:])
			return rx, tx
		}
	}
	
	// Single value - use for both
	v := ParseInt64(s)
	return v, v
}
