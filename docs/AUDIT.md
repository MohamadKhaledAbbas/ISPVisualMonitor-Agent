# Auditing the Agent

This guide explains how to audit the ISP Visual Monitor Agent to verify exactly what data it collects and transmits.

## 🎯 Why Audit?

As an ISP, you need **complete confidence** that monitoring tools:
1. Only collect the data they claim to collect
2. Don't include hidden telemetry or backdoors
3. Transmit data securely to the intended destination
4. Don't leak customer information

This agent is open-source specifically to enable this verification.

## 📋 Audit Checklist

- [ ] Review source code for data collection logic
- [ ] Inspect protocol definitions for transmitted fields
- [ ] Run agent with audit logging enabled
- [ ] Capture and inspect network traffic
- [ ] Verify TLS certificates and encryption
- [ ] Test data redaction features
- [ ] Review dependencies for security

## 🔍 Source Code Review

### 1. Clone and Inspect Repository

```bash
git clone https://github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent.git
cd ISPVisualMonitor-Agent
```

### 2. Review Data Collection Logic

**Key files to review**:

```bash
# Collector interface
cat internal/collector/collector.go

# MikroTik collector implementation
cat internal/collector/mikrotik/collector.go

# What data structures are collected
cat pkg/models/models.go

# Privacy and redaction logic
cat internal/privacy/redactor.go
cat internal/privacy/audit.go
```

### 3. Check Protocol Definitions

Protocol buffers define EXACTLY what can be transmitted:

```bash
# View all proto definitions
ls -l api/proto/*.proto

# Metrics definition
cat api/proto/metrics.proto

# Sessions definition (customer data)
cat api/proto/sessions.proto

# Agent protocol
cat api/proto/agent.proto
```

**Key question**: Do these proto files define ANY fields you weren't expecting?

### 4. Search for Suspicious Patterns

```bash
# Search for potential data exfiltration
grep -r "http.Post\|http.Get" .
grep -r "net.Dial" .
grep -r "os.Execute" .

# Search for file operations
grep -r "os.Create\|os.WriteFile" internal/

# Search for external commands
grep -r "exec.Command" .
```

## 🧪 Runtime Auditing

### 1. Enable Audit Logging

Create test configuration:

```yaml
# test-audit.yaml
agent:
  id: "audit-test"
  name: "Audit Test Agent"

server:
  address: "localhost:50051"
  tls:
    enabled: false  # For easier inspection

license:
  key: "test-key"
  validation_url: "localhost:50052"

collection:
  interval_seconds: 60

routers:
  - id: "test-router"
    name: "Test Router"
    type: "mikrotik"
    address: "192.168.1.1"
    credentials:
      username: "admin"
      password: "admin"
    collect:
      system: true
      interfaces: true
      pppoe_sessions: true
      nat_sessions: false
      dhcp_leases: true

privacy:
  audit_logging: true
  audit_log_path: "./audit.log"  # Local file for review

logging:
  level: "debug"
  format: "json"
  output: "stdout"
```

### 2. Run Agent and Collect Audit Data

```bash
# Build agent
make build

# Run with audit config
./bin/ispagent --config test-audit.yaml

# Let it run for a few minutes, then stop (Ctrl+C)
```

### 3. Inspect Audit Log

```bash
# View audit log
cat audit.log

# Pretty print JSON
cat audit.log | jq '.'

# Filter specific event types
cat audit.log | jq 'select(.event_type == "data_collection")'

# Count collections per data type
cat audit.log | jq -r '.data_type' | sort | uniq -c
```

**Example audit entry**:
```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "event_type": "data_collection",
  "router_id": "test-router",
  "data_type": "system_metrics",
  "record_count": 1,
  "details": {
    "cpu_percent": 25.5,
    "interfaces": 2
  }
}
```

**Questions to ask**:
- Is the `data_type` what you expected?
- Are the `details` limited to what you configured?
- Is there any unexpected data collection?

## 🌐 Network Traffic Analysis

### 1. Capture Traffic with tcpdump

```bash
# Capture gRPC traffic (TLS disabled for inspection)
sudo tcpdump -i any -nn port 50051 -w agent-traffic.pcap

# Run agent in another terminal
./bin/ispagent --config test-audit.yaml

# Stop capture after a few minutes (Ctrl+C)
```

### 2. Analyze with Wireshark

```bash
wireshark agent-traffic.pcap
```

**Filters**:
- `tcp.port == 50051` - All gRPC traffic
- `grpc` - gRPC protocol messages (if Wireshark has gRPC dissector)

### 3. Analyze with tshark

```bash
# Summary of packets
tshark -r agent-traffic.pcap

# Extract HTTP/2 headers (gRPC uses HTTP/2)
tshark -r agent-traffic.pcap -Y "http2" -T fields -e http2.header.name -e http2.header.value

# Extract data streams
tshark -r agent-traffic.pcap -Y "http2.data.data" -T fields -e http2.data.data
```

### 4. Inspect TLS Certificates (Production)

With TLS enabled:

```bash
# Check certificate
openssl s_client -connect monitor.example.com:443 -showcerts

# Verify certificate matches your server
openssl s_client -connect monitor.example.com:443 -servername monitor.example.com < /dev/null | openssl x509 -noout -text
```

## 🔬 Verify Data Redaction

### 1. Test Without Redaction

```yaml
privacy:
  redact_usernames: false
  redact_ip_addresses: false
```

Check audit log for actual values.

### 2. Test With Redaction

```yaml
privacy:
  redact_usernames: true
  redact_ip_addresses: true
```

**Verify**:
- Usernames are hashed (e.g., `a7f3c891b4e6d2f9`)
- IPs are masked (e.g., `192.168.xxx.xxx`)

```bash
# Search for non-redacted IPs (should be none with redaction enabled)
cat audit.log | grep -oE '\b([0-9]{1,3}\.){3}[0-9]{1,3}\b' | grep -v 'xxx'
```

## 🔐 Security Audit

### 1. Check Dependencies

```bash
# View all dependencies
go list -m all

# Check for known vulnerabilities
go list -json -m all | nancy sleuth
```

### 2. Static Analysis

```bash
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run security scan
gosec ./...
```

### 3. Check for Hardcoded Secrets

```bash
# Search for potential secrets
grep -rE "(password|secret|key|token).*=.*['\"]" . --include="*.go"

# Use gitleaks
docker run -v $(pwd):/path zricethezav/gitleaks:latest detect --source="/path" -v
```

## 📊 Comparison Audit

### Before/After Data Collection

1. **Export router state before agent collection**:
```bash
ssh admin@192.168.1.1 "export show-sensitive"
```

2. **Let agent collect data**

3. **Export router state after**:
```bash
ssh admin@192.168.1.1 "export show-sensitive"
```

4. **Compare** - Configuration should be identical (read-only access)

## ✅ Audit Report Template

```markdown
# Agent Audit Report

**Date**: 2024-01-15
**Auditor**: John Doe
**Agent Version**: 1.0.0

## Source Code Review
- [ ] Reviewed data collection logic in collectors
- [ ] Verified protocol definitions match documentation
- [ ] Searched for suspicious patterns (none found / issues found)

## Runtime Audit
- [ ] Enabled audit logging and reviewed output
- [ ] Confirmed only configured data types are collected
- [ ] Verified no unexpected data collection

## Network Analysis
- [ ] Captured and analyzed network traffic
- [ ] Verified TLS encryption is working
- [ ] Confirmed data transmission matches audit log

## Security
- [ ] Scanned dependencies for vulnerabilities
- [ ] Ran static security analysis
- [ ] No hardcoded secrets found

## Data Redaction
- [ ] Tested username redaction (working / not working)
- [ ] Tested IP address redaction (working / not working)

## Findings
[List any concerns or observations]

## Recommendation
[ ] Approved for deployment
[ ] Requires changes before deployment
[ ] Not approved

## Notes
[Additional comments]
```

## 🚨 Red Flags to Watch For

- ❌ Network connections to unexpected destinations
- ❌ Files created outside configured paths
- ❌ Data collection when all collectors are disabled
- ❌ Unencrypted transmission of sensitive data
- ❌ Hardcoded credentials or API keys
- ❌ Suspicious dependencies or modules
- ❌ Code obfuscation or minification

## 📞 Reporting Concerns

If you find something suspicious:

1. **Document thoroughly**: Screenshots, logs, packet captures
2. **Report privately**: security@ispmonitor.com
3. **Do not deploy** until resolved

## 🔄 Regular Re-Auditing

**Best practice**: Re-audit on every major version update.

```bash
# Check for updates
git fetch --tags
git log --oneline v1.0.0..v1.1.0

# Review changes
git diff v1.0.0..v1.1.0

# Look for changes in critical files
git diff v1.0.0..v1.1.0 -- internal/collector/
git diff v1.0.0..v1.1.0 -- api/proto/
```

## 📚 Additional Resources

- [PRIVACY.md](PRIVACY.md) - What data is collected
- [CONFIGURATION.md](CONFIGURATION.md) - How to configure collection
- [OWASP Code Review Guide](https://owasp.org/www-pdf-archive/OWASP_Code_Review_Guide_v2.pdf)

---

**Remember**: Trust, but verify. This agent is open-source for exactly this reason.
