# Privacy & Data Collection

This document provides complete transparency about what data the ISP Visual Monitor Agent collects from your infrastructure.

## 🎯 Philosophy

We believe ISPs should have **complete visibility** into what monitoring tools collect from their networks. This agent is open-source specifically to enable auditing before deployment.

## 📊 Data Collected

### 1. System Metrics

**What**: Basic router health information

**Fields Collected**:
- `cpu_percent` - CPU utilization percentage (0-100)
- `memory_percent` - Memory utilization percentage (0-100)
- `memory_total_bytes` - Total installed memory
- `memory_used_bytes` - Currently used memory
- `uptime_seconds` - Time since last reboot
- `temperature_celsius` - Board temperature (if available)
- `firmware_version` - RouterOS/firmware version string
- `board_name` - Hardware model identifier

**Why**: Essential for capacity planning and proactive maintenance.

**Privacy Impact**: ✅ No PII - Only hardware and performance metrics.

### 2. Interface Statistics

**What**: Network interface traffic and error counters

**Fields Collected**:
- `name` - Interface name (e.g., "ether1", "sfp-sfpplus1")
- `description` - Interface description/label
- `is_up` - Link status (true/false)
- `speed_mbps` - Link speed in Mbps
- `rx_bytes` - Received bytes counter
- `tx_bytes` - Transmitted bytes counter
- `rx_packets` - Received packets counter
- `tx_packets` - Transmitted packets counter
- `rx_errors` - Receive error counter
- `tx_errors` - Transmit error counter
- `rx_drops` - Receive drop counter
- `tx_drops` - Transmit drop counter

**Why**: Identify congestion, errors, and capacity issues.

**Privacy Impact**: ✅ No PII - Only interface identifiers and counters.

### 3. PPPoE Sessions

**What**: Active customer sessions (when `pppoe_sessions: true`)

**Fields Collected**:
- `session_id` - Internal session identifier
- `username` - PPPoE username (⚠️ **can be redacted**)
- `calling_station_id` - Customer MAC address
- `framed_ip` - Assigned IP address (⚠️ **can be redacted**)
- `session_time_seconds` - Duration of current session
- `bytes_in` - Downloaded bytes
- `bytes_out` - Uploaded bytes
- `status` - Session state (active, disconnected, etc.)
- `connect_time` - Session start timestamp

**Why**: Monitor customer connectivity, usage patterns, and service quality.

**Privacy Impact**: ⚠️ **Contains customer identifiers**

**Redaction Options**:
```yaml
privacy:
  redact_usernames: true    # Hashes usernames
  redact_ip_addresses: true # Masks IPs to xxx.xxx.0.0
```

### 4. NAT Sessions (Optional)

**What**: Active NAT connection tracking (when `nat_sessions: true`)

**Fields Collected**:
- `protocol` - TCP/UDP/ICMP
- `src_address` - Internal source IP (⚠️ **can be redacted**)
- `src_port` - Internal source port
- `dst_address` - External destination IP
- `dst_port` - External destination port
- `translated_address` - Public NAT IP
- `translated_port` - Public NAT port
- `bytes` - Total bytes transferred
- `packets` - Total packets transferred

**Why**: Capacity planning, troubleshooting connection issues.

**Privacy Impact**: ⚠️ **Contains customer IPs and browsing destinations**

**Default**: Disabled by default due to privacy concerns.

### 5. DHCP Leases

**What**: DHCP IP address assignments (when `dhcp_leases: true`)

**Fields Collected**:
- `mac_address` - Client MAC address
- `ip_address` - Assigned IP (⚠️ **can be redacted**)
- `hostname` - Client-provided hostname
- `lease_start` - Lease start time
- `lease_end` - Lease expiration time
- `status` - Lease status (bound, expired, etc.)

**Why**: Track IP allocation, identify IP exhaustion.

**Privacy Impact**: ⚠️ **Contains customer device identifiers**

## 🔍 Audit Logging

When `privacy.audit_logging: true`, every data collection event is logged locally:

```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "event_type": "data_collection",
  "router_id": "router-01",
  "data_type": "pppoe_sessions",
  "record_count": 47,
  "details": {
    "router_name": "Core Router",
    "collection_duration_ms": 234
  }
}
```

**Audit log location**: `/var/log/ispagent/audit.log` (configurable)

## 🔒 Data Redaction

### Username Redaction

When enabled, converts usernames to one-way hashes:

```
Original:   customer@example.com
Redacted:   a7f3c891b4e6d2f9
```

**Algorithm**: SHA-256 hash (first 16 chars)

### IP Address Redaction

When enabled, masks host portion of IPs:

```
IPv4:       192.168.1.100  →  192.168.xxx.xxx
IPv6:       2001:db8::1     →  2001:db8::xxxx
```

### Configuration

```yaml
privacy:
  audit_logging: true              # Enable local audit trail
  audit_log_path: "/var/log/ispagent/audit.log"
  redact_usernames: true           # Hash customer usernames
  redact_ip_addresses: true        # Mask IP addresses
```

## 📡 Data Transmission

All data is transmitted to the ISP Monitor server via:

- **Protocol**: gRPC over TLS 1.2+
- **Authentication**: API key per agent
- **Frequency**: Configurable (default: 60 seconds)
- **Destination**: Your configured server endpoint

### Server Communication Audit

Every transmission is also logged:

```json
{
  "timestamp": "2024-01-15T10:31:00Z",
  "event_type": "data_transmission",
  "router_id": "router-01",
  "data_type": "metrics",
  "record_count": 1,
  "details": {
    "destination": "monitor.example.com:443",
    "transmission_duration_ms": 45
  }
}
```

## 🛡️ Data Retention

- **Agent Side**: Audit logs are rotated locally (your choice)
- **Server Side**: Configured in ISP Monitor server settings
- **Default Server Retention**: 90 days (configurable per tier)

## 🚫 What We DON'T Collect

- ✅ Packet payloads or deep packet inspection
- ✅ Web browsing history or URLs
- ✅ Email content or metadata
- ✅ VPN session content
- ✅ Customer passwords or credentials
- ✅ DNS query contents (only aggregate stats if enabled)

## 🔍 How to Audit Before Deployment

### 1. Review the Source Code

```bash
# Clone repository
git clone https://github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent.git
cd ISPVisualMonitor-Agent

# Review data collection logic
cat internal/collector/mikrotik/collector.go
```

### 2. Inspect Protocol Definitions

```bash
# Review what fields are defined in protobuf
cat api/proto/metrics.proto
cat api/proto/sessions.proto
```

### 3. Enable Audit Logging

```yaml
privacy:
  audit_logging: true
  audit_log_path: "/tmp/audit.log"
```

Then run the agent and inspect `/tmp/audit.log` to see exactly what was collected.

### 4. Network Capture

Use tcpdump or Wireshark to inspect gRPC traffic (if TLS is disabled for testing):

```bash
sudo tcpdump -i any -nn port 50051 -w /tmp/agent-traffic.pcap
```

## 📞 Questions?

If you have privacy concerns or questions:

1. **Open an Issue**: [GitHub Issues](https://github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/issues)
2. **Email**: privacy@ispmonitor.com
3. **Review Code**: This is open-source for this exact reason!

## 📜 Compliance

This agent is designed to help ISPs comply with:

- **GDPR**: Customer data minimization and transparency
- **CCPA**: Consumer privacy rights
- **Industry Standards**: Network monitoring best practices

**Note**: You are responsible for ensuring your deployment complies with applicable regulations. We provide the tools for transparency and control.

---

**Last Updated**: 2024-01-15  
**Version**: 1.0.0
