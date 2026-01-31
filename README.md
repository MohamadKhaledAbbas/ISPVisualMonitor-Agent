# ISP Visual Monitor - Agent

[![CI](https://github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/workflows/CI/badge.svg)](https://github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/actions)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent)](https://goreportcard.com/report/github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent)

**Open-source monitoring agent for Internet Service Providers**

This is the data collection agent for [ISP Visual Monitor](https://github.com/MohamadKhaledAbbas/ISPVisualMonitor) - a complete ISP monitoring and management platform. The agent is distributed as a separate repository to provide **full transparency** to ISPs about exactly what data is being collected from their infrastructure.

## 🎯 Key Features

- **🔍 Fully Transparent**: Open-source so ISPs can audit exactly what data is collected
- **🔒 Privacy-Focused**: All collected data is logged locally with optional redaction
- **⚡ Lightweight**: Designed to run on routers, VMs, or dedicated monitoring boxes
- **🔐 Secure**: gRPC with TLS, API key authentication, and hardware-based licensing
- **🌐 Flexible Deployment**: Supports both cloud and on-premise modes
- **📊 Vendor Support**: MikroTik RouterOS (with Cisco/Juniper planned)

## 🏗️ Architecture

```
┌─────────────┐
│   Router    │ ◄──── RouterOS API
└──────┬──────┘
       │
       ▼
┌─────────────┐       gRPC/TLS      ┌──────────────┐
│ISP Agent    ├───────────────────► │ ISP Monitor  │
│(This Repo)  │                      │   Server     │
└─────────────┘                      └──────────────┘
       │
       ▼
┌─────────────┐
│ Audit Log   │ ◄──── Local audit trail
└─────────────┘
```

## 📦 What This Agent Collects

See [PRIVACY.md](docs/PRIVACY.md) for a complete, detailed breakdown. Summary:

- **System Metrics**: CPU, memory, uptime, temperature
- **Interface Statistics**: Traffic counters, errors, drops
- **PPPoE Sessions**: Username (hashed if configured), session time, bandwidth
- **NAT Sessions**: Connection tracking for capacity planning
- **DHCP Leases**: IP allocation tracking

**All data collection is logged locally for audit purposes.**

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or later (for building from source)
- MikroTik RouterOS device with API access
- Valid ISP Monitor license key

### Installation

#### Option 1: Pre-built Binary

```bash
# Download latest release
curl -LO https://github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/releases/latest/download/ispagent-linux-amd64

# Make executable
chmod +x ispagent-linux-amd64

# Move to system path
sudo mv ispagent-linux-amd64 /usr/local/bin/ispagent

# Create config directory
sudo mkdir -p /etc/ispagent

# Copy example config
sudo curl -o /etc/ispagent/agent.yaml https://raw.githubusercontent.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/main/configs/agent.yaml.example
```

#### Option 2: Build from Source

```bash
# Clone repository
git clone https://github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent.git
cd ISPVisualMonitor-Agent

# Build
make build

# Install
sudo make install
```

### Configuration

Edit `/etc/ispagent/agent.yaml`:

```yaml
agent:
  name: "my-agent"

server:
  address: "monitor.yourcompany.com:443"
  tls:
    enabled: true
    ca_cert: "/etc/ispagent/ca.crt"

license:
  key: "your-license-key-here"

routers:
  - id: "router-01"
    name: "Core Router"
    type: "mikrotik"
    address: "192.168.1.1"
    credentials:
      username: "admin"
      password: "your-password"
    collect:
      system: true
      interfaces: true
      pppoe_sessions: true
```

See [CONFIGURATION.md](docs/CONFIGURATION.md) for all options.

### Running

```bash
# Start agent
ispagent --config /etc/ispagent/agent.yaml

# Or use systemd (recommended)
sudo systemctl enable ispagent
sudo systemctl start ispagent
sudo systemctl status ispagent
```

## 📚 Documentation

- **[Privacy & Data Collection](docs/PRIVACY.md)** - What data is collected and why
- **[Configuration Reference](docs/CONFIGURATION.md)** - All configuration options
- **[Building from Source](docs/BUILDING.md)** - Build instructions and dependencies
- **[Auditing the Agent](docs/AUDIT.md)** - How to audit data before transmission
- **[Main Project](https://github.com/MohamadKhaledAbbas/ISPVisualMonitor)** - Server and web UI

## 🔐 Security

- **TLS 1.2+** for all server communication
- **API Key Authentication** per agent
- **Hardware Fingerprinting** for license validation
- **Audit Logging** of all data collection activities
- **Optional Data Redaction** (usernames, IP addresses)

See [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting PRs.

### Development

```bash
# Run tests
make test

# Run linter
make lint

# Build all platforms
make build-all
```

## 📄 License

This project is licensed under the Apache License 2.0 - see [LICENSE](LICENSE) file.

**Note**: While the agent is open-source, the ISP Monitor server requires a commercial license. Contact us for pricing.

## 🆘 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/issues)
- **Commercial Support**: support@ispmonitor.com

## 🗺️ Roadmap

- [x] MikroTik RouterOS support
- [ ] Cisco IOS/IOS-XE support
- [ ] Juniper JunOS support
- [ ] SNMP fallback collector
- [ ] NetFlow/IPFIX collection
- [ ] Webhooks for alerts

## 🙏 Acknowledgments

Built with:
- [gRPC](https://grpc.io/) for efficient communication
- [Protocol Buffers](https://protobuf.dev/) for typed data structures
- Go standard library for reliability

---

**Made with ❤️ for ISPs who value transparency and privacy**
