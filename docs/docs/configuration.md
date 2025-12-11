## DNS Settings

### Network Configuration

`dns.address`

The IP address on which the DNS server will listen for incoming queries.

**Default:** `0.0.0.0` (all IPv4 addresses)

!!! tip "Binding Options"

    Use `0.0.0.0` to listen on all interfaces, or specify a particular IP for more restrictive binding.

---

`dns.gateway`

Gateway address used when performing local lookups, primarily for finding hostnames of local clients. Will be populated if not set upon first startup.

**Default:** `192.168.0.1:53` (example)

---

### Performance & Caching

`dns.cacheTTL`

Maximum time (in seconds) to keep resolved domains in cache. The server uses either this value or the DNS response TTL, whichever is smaller.

**Default:** `360` seconds (6 minutes)

!!! info "Cache Behavior"

    Lower values provide more up-to-date information but may result in fewer cached responses and increased upstream queries.

---

`dns.udpSize`

UDP buffer size for incoming DNS queries in bytes. This follows the standard DNS-over-UDP packet size limit per RFC 1035.

**Default:** `512`

---

### Ports

UDP / TCP

`dns.ports.udptcp`

Port for standard DNS queries. The server listens on both UDP and TCP on this port.

**Default:** `53`

---

DNS-over-TLS (DoT)

`dns.ports.dot`

Port for DNS-over-TLS encrypted queries.

**Default:** `853`

---

DNS-over-HTTPS (DoH)

`dns.ports.doh`

Port for DNS-over-HTTPS encrypted queries.

**Default:** `443`

---

### TLS Configuration

!!! warning "TLS Setup Required"

    DoT and DoH servers will not start unless valid TLS certificates are configured.

`dns.tls.enabled`

Enable or disable TLS functionality for DoT and DoH.

**Default:** `false`

`dns.tls.cert`

Path to the TLS certificate file in PEM format.

**Default:** `""` (empty)

`dns.tls.key`

Path to the TLS private key file.

**Default:** `""` (empty)

---

### Upstream DNS Servers

`dns.upstream.preferred`

Primary DNS server to forward queries to.

**Default:** `8.8.8.8:53` (Google DNS)

`dns.upstream.fallback`

List of backup DNS servers used if the primary server fails.

**Default:** `[1.1.1.1:53]` (Cloudflare DNS)

!!! example "Multiple Fallbacks"

    ```yaml
    dns:
      upstream:
        preferred: 8.8.8.8:53
        fallback:
          - 1.1.1.1:53
          - 9.9.9.9:53
    ```

---

## API & Web Interface

### Server Configuration

`api.port`

Port for accessing the dashboard and API endpoints.

**Default:** `8080`

!!! info "Accessing the Dashboard"

    Navigate to `http://your-server-ip:8080` in your browser.

---

### Authentication

!!! warning "Production Security"

    Always enable authentication in production environments!

`api.authentication`

Controls whether login is required to access the dashboard.

**Default:** `true`

!!! note "First Startup"

    An admin account is created automatically on first startup. **Check the logs for the generated password.**

---

### JWT Secret

`api.jwtSecret`

Secret key used for signing JWT tokens.
If empty, a random key will be generated automatically.

**Default:** `""` (empty)

### Rate Limiting

`api.ratelimit.enabled`

Enable or disable rate limiting (currently protects only the login route).

**Default:** `false`

`api.ratelimit.maxTries`

Maximum number of requests before rate limiting activates.

**Default:** `5` attempts

`api.ratelimit.window`

Duration in minutes that rate limiting remains active after the limit is reached.

**Default:** `5` minutes

---

## Logging

`logging.enabled`

Master toggle for all logging functionality.

**Default:** `true`

!!! tip "Privacy & Performance"

    Disable logging for privacy-focused deployments or to reduce disk I/O.

`logging.level`

Controls the severity of log messages displayed. Each level includes all higher-numbered levels.

**Default:** `1` (Info)

| Level | Name    | Description                                  |
| ----- | ------- | -------------------------------------------- |
| 0     | Debug   | Most verbose, includes all messages          |
| 1     | Info    | Normal operation messages                    |
| 2     | Warning | Potential issues that don't affect operation |
| 3     | Error   | Serious problems only                        |

---

## Miscellaneous Settings

### Application Updates

`misc.inAppUpdate`

Enables or disables the built-in update functionality.

**Default:** `false`

!!! info "Update Behavior by Deployment Type"

    | Deployment | Setting | Behavior                                                                                              |
    |------------|---------|-------------------------------------------------------------------------------------------------------|
    | Docker     | `false` | Manual updates: stop container, remove, pull new image                                                |
    | Docker     | `true`  | Dashboard updater fetches latest binary and restarts container automatically                          |
    | Standalone | `false` | Manual updates via installer or `updater.sh`                                                          |
    | Standalone | `true`  | Dashboard updater installs new binary (manual restart required)                                       |

---

### Data Retention

`misc.statisticsRetention`

Number of days to retain statistics and query logs.

**Default:** `7` days

!!! tip "Storage Optimization"

    Lower values save disk space but provide less historical data for analysis.

---

### Dashboard Serving

`misc.dashboard`

Controls whether the web dashboard UI is served.

**Default:** `true`

!!! note "API-Only Mode"

    When set to `false`, the API remains available but the dashboard won't be served. Useful for headless deployments.

---

### Blacklist Management

`misc.scheduledBlacklistUpdates`

Enable automatic daily updates for blacklists at midnight.

**Default:** `true`

!!! success "Recommended"

    Keep this enabled to ensure your blacklists stay current with the latest threat intelligence.

---

## Quick Start Example

This is the default configuration that will be generated unless another config already exists.

```yaml
dns:
  address: 0.0.0.0
  gateway: 192.168.0.1:53
  cacheTTL: 3600
  udpSize: 512
  tls:
    enabled: false
    cert: ""
    key: ""
  upstream:
    preferred: 8.8.8.8:53
    fallback:
      - 1.1.1.1:53
  ports:
    udptcp: 53
    dot: 853
    doh: 443
api:
  port: 8080
  authentication: true
  jwtSecret: ""
  rateLimit:
    enabled: true
    maxTries: 5
    window: 5
logging:
  enabled: true
  level: 1
misc:
  inAppUpdate: false
  statisticsRetention: 7
  dashboard: true
  scheduledBlacklistUpdates: true
```
