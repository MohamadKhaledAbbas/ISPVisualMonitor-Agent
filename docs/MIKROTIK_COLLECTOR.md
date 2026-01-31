# MikroTik RouterOS Collector

The MikroTik collector provides comprehensive monitoring capabilities for MikroTik RouterOS devices. It uses the native RouterOS API protocol to collect system metrics, interface statistics, PPPoE sessions, NAT connections, and DHCP leases.

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Configuration](#configuration)
- [Metrics Collected](#metrics-collected)
- [RouterOS Setup](#routeros-setup)
- [Security Best Practices](#security-best-practices)
- [Troubleshooting](#troubleshooting)
- [API Protocol Details](#api-protocol-details)

## Features

- **Full RouterOS API Support**: Native API protocol implementation (sentence-based)
- **TLS Support**: Secure connections via port 8729
- **Dual Authentication**: Challenge-response (RouterOS <6.43) and new login method (6.43+)
- **Circuit Breaker**: Automatic protection against connection storms
- **Rate Calculations**: Per-interface traffic rate calculations
- **Interface Filtering**: Include/exclude patterns for selective monitoring
- **NAT Sampling**: Configurable sampling for high-traffic routers
- **Privacy Compliant**: Integration with audit logging and data redaction

## Requirements

### RouterOS Version
- RouterOS 6.x or 7.x (6.43+ recommended for new login method)

### RouterOS Services
- API service enabled (port 8728) or API-SSL service enabled (port 8729)

### User Permissions
- `read` - Read access to router configuration
- `api` - API access permission

## Configuration

### Basic Configuration

```yaml
routers:
  - id: "router-01"
    name: "Core Router"
    type: "mikrotik"
    address: "192.168.1.1"
    credentials:
      username: "${MIKROTIK_USER}"
      password: "${MIKROTIK_PASS}"
```

### Full Configuration

```yaml
routers:
  - id: "router-01"
    name: "Core Router"
    type: "mikrotik"
    address: "192.168.1.1"
    credentials:
      username: "${MIKROTIK_USER}"
      password: "${MIKROTIK_PASS}"
    metadata:
      api:
        port: 8729
        use_tls: true
        insecure_skip_verify: false  # Set to true only for self-signed certs (not recommended for production)
        timeout: 10s
        retry_attempts: 3
        retry_delay: 1s
      collect:
        system: true
        interfaces: true
        pppoe: true
        nat: false  # Disabled by default - expensive operation
        dhcp: true
      interface_include:
        - "ether*"
        - "sfp*"
      interface_exclude:
        - "bridge-local"
      nat:
        sampling_enabled: true
        sample_rate: 0.1  # Sample 10% of connections
        max_connections: 10000
```

### Environment Variables

The collector supports credential injection via environment variables:

```bash
export MIKROTIK_USER=ispmonitor
export MIKROTIK_PASS=secure-password
```

## Metrics Collected

### System Metrics

| Metric | Description | RouterOS Command |
|--------|-------------|------------------|
| `cpu_percent` | CPU utilization | `/system/resource/print` |
| `memory_percent` | Memory utilization | `/system/resource/print` |
| `memory_total_bytes` | Total memory | `/system/resource/print` |
| `memory_used_bytes` | Used memory | Calculated |
| `uptime_seconds` | System uptime | `/system/resource/print` |
| `temperature_celsius` | CPU/system temperature | `/system/health/print` |
| `firmware_version` | RouterOS version | `/system/resource/print` |
| `board_name` | Hardware model | `/system/resource/print` |
| `architecture` | CPU architecture | `/system/resource/print` |
| `license_level` | RouterOS license level | `/system/license/print` |
| `disk_total_bytes` | Total storage | `/system/resource/print` |
| `disk_used_bytes` | Used storage | Calculated |
| `voltage_mv` | Input voltage (if available) | `/system/health/print` |
| `fan_speed_rpm` | Fan speed (if available) | `/system/health/print` |

### Interface Metrics

| Metric | Description | RouterOS Command |
|--------|-------------|------------------|
| `name` | Interface name | `/interface/print` |
| `type` | Interface type (ether, bridge, etc.) | `/interface/print` |
| `is_up` | Running state | `/interface/print` |
| `enabled` | Administrative state | `/interface/print` |
| `rx_bytes` | Received bytes | `/interface/print` |
| `tx_bytes` | Transmitted bytes | `/interface/print` |
| `rx_packets` | Received packets | `/interface/print` |
| `tx_packets` | Transmitted packets | `/interface/print` |
| `rx_errors` | Receive errors | `/interface/print` |
| `tx_errors` | Transmit errors | `/interface/print` |
| `rx_drops` | Receive drops | `/interface/print` |
| `tx_drops` | Transmit drops | `/interface/print` |
| `speed_mbps` | Link speed (Ethernet only) | `/interface/ethernet/print` |
| `full_duplex` | Duplex mode (Ethernet only) | `/interface/ethernet/print` |
| `mtu` | Maximum transmission unit | `/interface/print` |
| `mac_address` | MAC address | `/interface/print` |
| `rx_bytes_per_sec` | Calculated RX rate | Derived |
| `tx_bytes_per_sec` | Calculated TX rate | Derived |

### PPPoE Sessions

| Metric | Description | RouterOS Command |
|--------|-------------|------------------|
| `username` | PPPoE username | `/ppp/active/print` |
| `caller_id` | Client MAC address | `/ppp/active/print` |
| `address` | Assigned IP address | `/ppp/active/print` |
| `uptime_seconds` | Session duration | `/ppp/active/print` |
| `rx_bytes` | Session received bytes | `/ppp/active/print` |
| `tx_bytes` | Session transmitted bytes | `/ppp/active/print` |
| `rate_limit` | Applied rate limit | `/ppp/active/print` |
| `service` | PPPoE service name | `/ppp/active/print` |

### NAT/Connection Tracking

| Metric | Description | RouterOS Command |
|--------|-------------|------------------|
| `total_connections` | Total active connections | Count |
| `tcp_connections` | TCP connections | Count by protocol |
| `udp_connections` | UDP connections | Count by protocol |
| `icmp_connections` | ICMP connections | Count by protocol |
| Connection details | Per-connection info | `/ip/firewall/connection/print` |

### DHCP Leases

| Metric | Description | RouterOS Command |
|--------|-------------|------------------|
| `address` | Leased IP address | `/ip/dhcp-server/lease/print` |
| `mac_address` | Client MAC | `/ip/dhcp-server/lease/print` |
| `hostname` | Client hostname | `/ip/dhcp-server/lease/print` |
| `status` | Lease status | `/ip/dhcp-server/lease/print` |
| `expires_after` | Time until expiry | `/ip/dhcp-server/lease/print` |
| Pool utilization | Pool usage statistics | Calculated |

## RouterOS Setup

### Creating a Monitoring User

```routeros
# Create a monitoring user group with minimal permissions
/user group add name=monitoring policy=read,api,!write,!policy,!reboot,!sensitive

# Create the monitoring user
/user add name=ispmonitor group=monitoring password="secure-password"
```

### Enabling API Service

For plain API (port 8728):
```routeros
/ip service set api disabled=no address=10.0.0.0/24
```

For API over TLS (port 8729, recommended):
```routeros
# Generate or import a certificate first
/ip service set api-ssl disabled=no address=10.0.0.0/24
```

### Restricting API Access

```routeros
# Limit API access to specific IP addresses
/ip service set api address=10.0.0.100/32,10.0.0.101/32

# Or use firewall rules
/ip firewall filter add chain=input protocol=tcp dst-port=8728,8729 \
    src-address=10.0.0.0/24 action=accept
/ip firewall filter add chain=input protocol=tcp dst-port=8728,8729 action=drop
```

## Security Best Practices

1. **Use TLS (port 8729)**: Always use API-SSL in production environments
2. **Verify Certificates**: Keep `insecure_skip_verify: false` unless using self-signed certs
3. **Minimal Permissions**: Use the `monitoring` group with only `read,api` permissions
4. **Strong Passwords**: Use environment variables for credentials, never hardcode
5. **IP Restrictions**: Limit API access to monitoring server IPs only
6. **Firewall Rules**: Add firewall rules as an additional layer of protection
7. **Audit Logging**: Enable audit logging to track all data collection events

**Note on Self-Signed Certificates**: If your RouterOS uses self-signed certificates and you must set `insecure_skip_verify: true`, ensure that network access to the router is strictly controlled via IP restrictions and firewall rules.

## Troubleshooting

### Connection Issues

**"Connection refused"**
- Verify API service is enabled: `/ip service print`
- Check firewall rules: `/ip firewall filter print`
- Verify port is correct (8728 for plain, 8729 for TLS)

**"Authentication failed"**
- Verify username and password
- Check user has `api` permission: `/user print`
- For RouterOS <6.43, ensure challenge-response works

**"Circuit breaker open"**
- Too many consecutive failures have occurred
- Wait 30 seconds for automatic reset
- Check router connectivity

### Performance Issues

**High CPU on router**
- Reduce collection frequency
- Disable NAT collection or enable sampling
- Use interface filtering to limit monitored interfaces

**Slow collection**
- Increase timeout setting
- Check network latency to router
- Consider collecting fewer metric types

### Data Issues

**Missing temperature/voltage**
- Not all RouterOS hardware supports health monitoring
- Check `/system/health/print` directly on router

**Zero rate values**
- Normal on first collection (no previous data)
- Counter wrap is handled automatically

## API Protocol Details

The MikroTik API uses a sentence-based protocol over TCP:

### Sentence Structure
```
/command/path
=attribute=value
=another=attribute
<empty word to end sentence>
```

### Word Encoding
Words are length-prefixed with variable-length encoding:
- 0x00-0x7F: 1 byte length
- 0x80-0x3FFF: 2 bytes (0x8000 + length)
- 0x4000-0x1FFFFF: 3 bytes
- 0x200000-0xFFFFFFF: 4 bytes
- 0x10000000+: 5 bytes

### Reply Types
- `!re` - Data reply (one record)
- `!done` - Command completed
- `!trap` - Error occurred
- `!fatal` - Fatal error (connection will close)

### Authentication Methods

**New Method (RouterOS 6.43+)**:
```
/login
=name=admin
=password=secret
```

**Legacy Challenge-Response**:
1. Send `/login`
2. Receive `=ret=challenge` in hex
3. Calculate: MD5(0x00 + password + decoded_challenge)
4. Send `/login` with `=name=user` and `=response=00+hex_hash`

## Support

For issues specific to this collector, please open an issue on the GitHub repository.

For RouterOS-specific questions, refer to the [MikroTik Wiki](https://wiki.mikrotik.com/wiki/Manual:API).
