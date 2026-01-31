package mikrotik

import (
	"context"
	"sync"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/collector/mikrotik/api"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/models"
)

// InterfaceMetrics contains extended interface metrics with rate calculations.
type InterfaceMetrics struct {
	models.InterfaceMetrics
	
	// Extended fields
	Type           string  `json:"type,omitempty"`
	MAC            string  `json:"mac,omitempty"`
	MTU            int64   `json:"mtu,omitempty"`
	Enabled        bool    `json:"enabled"`
	RxBytesPerSec  float64 `json:"rx_bytes_per_sec,omitempty"`
	TxBytesPerSec  float64 `json:"tx_bytes_per_sec,omitempty"`
	RxPktsPerSec   float64 `json:"rx_pkts_per_sec,omitempty"`
	TxPktsPerSec   float64 `json:"tx_pkts_per_sec,omitempty"`
	FullDuplex     bool    `json:"full_duplex,omitempty"`
}

// interfaceState stores previous counter values for rate calculation.
type interfaceState struct {
	rxBytes    uint64
	txBytes    uint64
	rxPackets  uint64
	txPackets  uint64
	timestamp  time.Time
}

// interfaceTracker tracks interface state for rate calculations.
type interfaceTracker struct {
	mu     sync.RWMutex
	states map[string]*interfaceState
}

func newInterfaceTracker() *interfaceTracker {
	return &interfaceTracker{
		states: make(map[string]*interfaceState),
	}
}

func (t *interfaceTracker) updateAndCalculateRates(name string, rxBytes, txBytes, rxPkts, txPkts uint64) (rxBps, txBps, rxPps, txPps float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	prev, exists := t.states[name]

	if exists && !prev.timestamp.IsZero() {
		elapsed := now.Sub(prev.timestamp).Seconds()
		if elapsed > 0 {
			// Calculate deltas, handling counter wrap
			rxDelta := Handle64BitCounterWrap(rxBytes, prev.rxBytes)
			txDelta := Handle64BitCounterWrap(txBytes, prev.txBytes)
			rxPktDelta := Handle64BitCounterWrap(rxPkts, prev.rxPackets)
			txPktDelta := Handle64BitCounterWrap(txPkts, prev.txPackets)

			rxBps = float64(rxDelta) / elapsed
			txBps = float64(txDelta) / elapsed
			rxPps = float64(rxPktDelta) / elapsed
			txPps = float64(txPktDelta) / elapsed
		}
	}

	// Store current state
	t.states[name] = &interfaceState{
		rxBytes:   rxBytes,
		txBytes:   txBytes,
		rxPackets: rxPkts,
		txPackets: txPkts,
		timestamp: now,
	}

	return
}

// collectInterfaces collects interface metrics from the router.
func (c *Collector) collectInterfaces(ctx context.Context, client *api.Client) ([]InterfaceMetrics, error) {
	// Get all interfaces
	interfaces, err := client.Run(ctx, "/interface/print", nil)
	if err != nil {
		return nil, err
	}

	var result []InterfaceMetrics

	for _, iface := range interfaces {
		name := iface["name"]
		
		// Apply include/exclude filters
		if !MatchFilter(name, c.config.InterfaceInclude, c.config.InterfaceExclude) {
			continue
		}

		metrics := InterfaceMetrics{
			InterfaceMetrics: models.InterfaceMetrics{
				Name:        name,
				Description: iface["comment"],
				IsUp:        iface["running"] == "true",
				RxBytes:     ParseInt64(iface["rx-byte"]),
				TxBytes:     ParseInt64(iface["tx-byte"]),
				RxPackets:   ParseInt64(iface["rx-packet"]),
				TxPackets:   ParseInt64(iface["tx-packet"]),
				RxErrors:    ParseInt64(iface["rx-error"]),
				TxErrors:    ParseInt64(iface["tx-error"]),
				RxDrops:     ParseInt64(iface["rx-drop"]),
				TxDrops:     ParseInt64(iface["tx-drop"]),
			},
			Type:    iface["type"],
			MAC:     iface["mac-address"],
			MTU:     ParseInt64(iface["mtu"]),
			Enabled: iface["disabled"] != "true",
		}

		// Get detailed stats for ethernet interfaces
		if metrics.Type == "ether" {
			etherStats, err := c.getEthernetStats(ctx, client, name)
			if err == nil {
				metrics.SpeedMbps = etherStats.speed
				metrics.FullDuplex = etherStats.fullDuplex
			}
		}

		// Calculate rates
		rxBps, txBps, rxPps, txPps := c.ifaceTracker.updateAndCalculateRates(
			name,
			uint64(metrics.RxBytes),
			uint64(metrics.TxBytes),
			uint64(metrics.RxPackets),
			uint64(metrics.TxPackets),
		)
		metrics.RxBytesPerSec = rxBps
		metrics.TxBytesPerSec = txBps
		metrics.RxPktsPerSec = rxPps
		metrics.TxPktsPerSec = txPps

		result = append(result, metrics)
	}

	return result, nil
}

type ethernetStats struct {
	speed      int64
	fullDuplex bool
}

func (c *Collector) getEthernetStats(ctx context.Context, client *api.Client, name string) (*ethernetStats, error) {
	// Get ethernet interface stats
	sentence := api.NewSentence("/interface/ethernet/print")
	sentence.AddQuery("name", name)
	
	results, err := client.Run(ctx, "/interface/ethernet/print", map[string]string{
		".proplist": "name,speed,full-duplex",
	})
	if err != nil {
		return nil, err
	}

	for _, r := range results {
		if r["name"] == name {
			return &ethernetStats{
				speed:      ParseSpeed(r["speed"]),
				fullDuplex: r["full-duplex"] == "true",
			}, nil
		}
	}

	return &ethernetStats{}, nil
}
