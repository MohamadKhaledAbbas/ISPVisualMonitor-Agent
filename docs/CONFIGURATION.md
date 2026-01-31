# Configuration Reference

Complete reference for ISP Visual Monitor Agent configuration.

## 📄 Configuration File

Default location: `/etc/ispagent/agent.yaml`

Override with: `ispagent --config /path/to/config.yaml`

## 🔧 Configuration Sections

### Agent Identity

```yaml
agent:
  id: ""              # Auto-generated if empty (hostname-based)
  name: "agent-01"    # Human-readable name for this agent
```

**Fields**:
- `id`: Unique identifier for this agent. Auto-generated as `agent-<hostname>` if empty.
- `name`: Display name shown in ISP Monitor dashboard.

### Server Connection

```yaml
server:
  address: "monitor.example.com:443"
  tls:
    enabled: true
    ca_cert: "/etc/ispagent/ca.crt"
    client_cert: "/etc/ispagent/client.crt"
    client_key: "/etc/ispagent/client.key"
```

**Fields**:
- `address`: gRPC server hostname:port
- `tls.enabled`: Use TLS encryption (recommended: true)
- `tls.ca_cert`: Path to CA certificate for server validation
- `tls.client_cert`: Path to client certificate (if using mutual TLS)
- `tls.client_key`: Path to client private key (if using mutual TLS)

**Development Mode**: Set `tls.enabled: false` for testing only.

### License Configuration

```yaml
license:
  key: "${LICENSE_KEY}"
  validation_url: "https://license.ispmonitor.com/v1/validate"
  offline_grace_hours: 72
```

**Fields**:
- `key`: Your ISP Monitor license key (supports env var substitution)
- `validation_url`: License validation server endpoint
- `offline_grace_hours`: Hours agent can run without online validation

**Environment Variables**: Use `${VAR_NAME}` syntax to inject from environment.

### Collection Settings

```yaml
collection:
  interval_seconds: 60
```

**Fields**:
- `interval_seconds`: How often to collect metrics from routers (default: 60)

**Recommendations**:
- **High-frequency monitoring**: 30 seconds
- **Standard monitoring**: 60 seconds
- **Low-frequency monitoring**: 300 seconds (5 minutes)

### Router Configuration

```yaml
routers:
  - id: "router-01"
    name: "Core Router"
    type: "mikrotik"
    address: "192.168.1.1"
    credentials:
      username: "${ROUTER_USER}"
      password: "${ROUTER_PASS}"
      ssh_key: ""  # Optional: SSH key path for key-based auth
    collect:
      system: true
      interfaces: true
      pppoe_sessions: true
      nat_sessions: false
      dhcp_leases: true
    metadata:
      location: "datacenter-1"
      tier: "core"
```

**Router Fields**:
- `id`: Unique identifier for this router
- `name`: Display name
- `type`: Router type (`mikrotik`, `cisco`, `juniper` - only mikrotik currently supported)
- `address`: IP address or hostname
- `credentials`: Authentication details (supports env var substitution)

**Collection Flags**:
- `system`: System metrics (CPU, memory, uptime, temperature)
- `interfaces`: Interface statistics (traffic, errors, drops)
- `pppoe_sessions`: PPPoE session data (⚠️ contains customer info)
- `nat_sessions`: NAT connection tracking (⚠️ privacy sensitive, disabled by default)
- `dhcp_leases`: DHCP lease information

**Metadata**: Optional key-value pairs for organization (shown in dashboard).

### Privacy & Audit

```yaml
privacy:
  audit_logging: true
  audit_log_path: "/var/log/ispagent/audit.log"
  redact_usernames: false
  redact_ip_addresses: false
```

**Fields**:
- `audit_logging`: Enable local audit trail of all collections
- `audit_log_path`: Where to write audit logs
- `redact_usernames`: Hash usernames before transmission
- `redact_ip_addresses`: Mask IP addresses before transmission

**Recommendation**: Always enable `audit_logging` for transparency.

See [PRIVACY.md](PRIVACY.md) for details on what data redaction does.

### Logging

```yaml
logging:
  level: "info"
  format: "json"
  output: "/var/log/ispagent/agent.log"
```

**Fields**:
- `level`: Log level (`debug`, `info`, `warn`, `error`)
- `format`: Log format (`json`, `text`)
- `output`: Log destination (file path or `stdout`)

**Log Levels**:
- `debug`: Verbose output for troubleshooting
- `info`: Standard operational messages (recommended)
- `warn`: Warning messages only
- `error`: Error messages only

## 🔐 Security Best Practices

### 1. Use Environment Variables for Secrets

Instead of:
```yaml
license:
  key: "my-secret-license-key"
```

Use:
```yaml
license:
  key: "${LICENSE_KEY}"
```

Then set environment:
```bash
export LICENSE_KEY="my-secret-license-key"
```

### 2. Restrict Configuration File Permissions

```bash
sudo chown root:root /etc/ispagent/agent.yaml
sudo chmod 600 /etc/ispagent/agent.yaml
```

### 3. Use TLS in Production

Always enable TLS for production deployments:
```yaml
server:
  tls:
    enabled: true
```

### 4. Enable Audit Logging

Always enable audit logging:
```yaml
privacy:
  audit_logging: true
```

## 📝 Example Configurations

### Minimal Configuration

```yaml
agent:
  name: "edge-router-01"

server:
  address: "monitor.example.com:443"
  tls:
    enabled: true
    ca_cert: "/etc/ispagent/ca.crt"

license:
  key: "${LICENSE_KEY}"

routers:
  - id: "router-01"
    name: "Edge Router"
    type: "mikrotik"
    address: "10.0.0.1"
    credentials:
      username: "${ROUTER_USER}"
      password: "${ROUTER_PASS}"
    collect:
      system: true
      interfaces: true
```

### Privacy-Focused Configuration

```yaml
agent:
  name: "privacy-agent"

server:
  address: "monitor.example.com:443"
  tls:
    enabled: true
    ca_cert: "/etc/ispagent/ca.crt"

license:
  key: "${LICENSE_KEY}"

routers:
  - id: "router-01"
    name: "Core Router"
    type: "mikrotik"
    address: "192.168.1.1"
    credentials:
      username: "${ROUTER_USER}"
      password: "${ROUTER_PASS}"
    collect:
      system: true
      interfaces: true
      pppoe_sessions: true
      nat_sessions: false      # Disabled for privacy
      dhcp_leases: false       # Disabled for privacy

privacy:
  audit_logging: true
  audit_log_path: "/var/log/ispagent/audit.log"
  redact_usernames: true       # Hash usernames
  redact_ip_addresses: true    # Mask IPs

logging:
  level: "info"
  format: "json"
  output: "/var/log/ispagent/agent.log"
```

### Multi-Router Configuration

```yaml
agent:
  name: "multi-agent"

server:
  address: "monitor.example.com:443"
  tls:
    enabled: true
    ca_cert: "/etc/ispagent/ca.crt"

license:
  key: "${LICENSE_KEY}"

collection:
  interval_seconds: 60

routers:
  - id: "core-01"
    name: "Core Router 1"
    type: "mikrotik"
    address: "10.0.0.1"
    credentials:
      username: "${ROUTER_USER}"
      password: "${ROUTER_PASS}"
    collect:
      system: true
      interfaces: true
      pppoe_sessions: true
    metadata:
      location: "datacenter-1"
      role: "core"

  - id: "core-02"
    name: "Core Router 2"
    type: "mikrotik"
    address: "10.0.0.2"
    credentials:
      username: "${ROUTER_USER}"
      password: "${ROUTER_PASS}"
    collect:
      system: true
      interfaces: true
      pppoe_sessions: true
    metadata:
      location: "datacenter-2"
      role: "core"

  - id: "edge-01"
    name: "Edge Router"
    type: "mikrotik"
    address: "10.1.0.1"
    credentials:
      username: "${ROUTER_USER}"
      password: "${ROUTER_PASS}"
    collect:
      system: true
      interfaces: true
    metadata:
      location: "pop-1"
      role: "edge"

privacy:
  audit_logging: true
  audit_log_path: "/var/log/ispagent/audit.log"

logging:
  level: "info"
  format: "json"
  output: "/var/log/ispagent/agent.log"
```

## 🔄 Reloading Configuration

The agent does **not** support hot-reload. To apply configuration changes:

```bash
sudo systemctl restart ispagent
```

Or if running manually:
```bash
# Stop the agent (Ctrl+C or kill)
# Then restart with new config
ispagent --config /etc/ispagent/agent.yaml
```

## ✅ Validating Configuration

Test your configuration before deploying:

```bash
# Dry-run (planned feature)
ispagent --config /etc/ispagent/agent.yaml --validate

# Check logs for validation errors
ispagent --config /etc/ispagent/agent.yaml
# Look for "Invalid configuration" messages
```

## 🆘 Troubleshooting

### "Failed to read config file"

- Check file path: `ls -l /etc/ispagent/agent.yaml`
- Check permissions: Should be readable by agent user
- Check YAML syntax: `yamllint /etc/ispagent/agent.yaml`

### "License validation failed"

- Verify license key is correct
- Check network connectivity to license server
- Review `offline_grace_hours` setting

### "Failed to connect to router"

- Verify router address is correct and reachable
- Check router credentials
- Ensure RouterOS API is enabled
- Check firewall rules allow agent → router communication

---

**Need Help?** See [README.md](../README.md) or open an issue.
