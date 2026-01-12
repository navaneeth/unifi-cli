# Unifi CLI

A command-line interface for managing Unifi Network devices using the official Unifi Network API.

## Features

- List connected clients with detailed information
- **Powerful filtering with SQL WHERE clause syntax**
- Simple filter flags for common use cases (wired/wireless, AP, blocked)
- Table and JSON output formats
- Support for self-signed certificates (insecure mode enabled by default)
- Configuration via environment variables, config file, or command-line flags

## Installation

### Using go install

```bash
make install
```

This will install the `unifi` binary to your `$GOPATH/bin` directory.

### Building from source

```bash
make build
```

The binary will be created at `./bin/unifi`.

## Configuration

The CLI can be configured using multiple methods (in order of precedence):

1. Command-line flags
2. Environment variables
3. Configuration file
4. Default values

### Environment Variables

- `UNIFI_HOST` - Unifi controller host (e.g., `https://unifi.example.com`)
- `UNIFI_API_KEY` - API key for authentication
- `UNIFI_SITE` - Site ID (default: `default`)

### Configuration File

Create a configuration file at `~/.unifi-cli.yaml`:

```yaml
host: https://unifi.example.com
api_key: your-api-key-here
site: default
insecure: true  # Skip TLS verification (default)
```

You can also specify a custom config file path using the `--config` flag.

### Command-line Flags

Global flags available for all commands:

- `--config, -c` - Path to config file
- `--host` - Unifi controller host
- `--site` - Site ID
- `--insecure, -k` - Skip TLS certificate verification (default: true)

## Usage

### List Connected Clients

List all currently connected clients:

```bash
unifi clients list
```

Output in JSON format:

```bash
unifi clients list --format json
```

or

```bash
unifi clients list -f json
```

### Examples

```bash
# Using environment variables
export UNIFI_HOST="https://unifi.example.com"
export UNIFI_API_KEY="your-api-key"
unifi clients list

# Using command-line flags
unifi --host https://unifi.example.com clients list

# Using a custom config file
unifi --config /path/to/config.yaml clients list

# Output as JSON
unifi clients list --format json
```

## Filtering Clients

The CLI supports powerful filtering capabilities to help you find specific clients.

### Simple Filter Flags

Use convenient flags for common filtering scenarios:

```bash
# Show only wired clients
unifi clients list --wired

# Show only wireless clients
unifi clients list --wireless

# Show only blocked clients
unifi clients list --blocked

# Filter by Access Point MAC address
unifi clients list --ap aa:bb:cc:dd:ee:ff
```

### SQL WHERE Clause Filtering

For advanced filtering, use the `--filter` flag with SQL WHERE clause syntax:

```bash
# Clients with good signal strength
unifi clients list --filter "signal >= -65"

# Clients on a specific SSID
unifi clients list --filter "essid = 'HomeWiFi'"

# Wireless clients with good signal
unifi clients list --filter "is_wired = 0 AND signal >= -60"

# Clients with SSID containing "Guest"
unifi clients list --filter "essid LIKE '%Guest%'"

# Clients with signal in a specific range
unifi clients list --filter "signal BETWEEN -70 AND -50"

# Clients by hostname
unifi clients list --filter "hostname IN ('iphone', 'ipad', 'macbook')"

# Complex queries with parentheses
unifi clients list --filter "(essid = 'HomeWiFi' OR essid = 'GuestWiFi') AND signal >= -65"
```

### Combining Filters

You can combine simple flags with SQL filters:

```bash
# Wireless clients with good signal
unifi clients list --wireless --filter "signal >= -65"

# Specific AP with hostname pattern
unifi clients list --ap aa:bb:cc:dd:ee:ff --filter "hostname LIKE '%phone%'"

# Wired clients only, output as JSON
unifi clients list --wired --format json
```

### Available Filter Fields

| Field | Type | Description |
|-------|------|-------------|
| `mac` | TEXT | Client MAC address |
| `name` | TEXT | User-assigned client name |
| `hostname` | TEXT | Client hostname |
| `ip` | TEXT | Client IP address |
| `is_wired` | INTEGER | 1 for wired, 0 for wireless |
| `blocked` | INTEGER | 1 if blocked, 0 otherwise |
| `essid` | TEXT | SSID (wireless clients only) |
| `ap_mac` | TEXT | Access Point MAC address |
| `signal` | INTEGER | Signal strength in dBm (negative values) |
| `uptime` | INTEGER | Uptime in seconds |
| `tx_rate` | INTEGER | Transmission rate in Mbps |
| `rx_rate` | INTEGER | Receive rate in Mbps |
| `satisfaction` | INTEGER | Client satisfaction score (0-100) |
| `channel` | INTEGER | WiFi channel |
| `rssi` | INTEGER | RSSI value |
| `sw_mac` | TEXT | Switch MAC address (wired clients) |
| `sw_port` | INTEGER | Switch port number (wired clients) |
| `tx_bytes` | INTEGER | Total transmitted bytes |
| `rx_bytes` | INTEGER | Total received bytes |

### SQL Operators Supported

- Comparison: `=`, `!=`, `>`, `<`, `>=`, `<=`
- Pattern matching: `LIKE` (use `%` as wildcard)
- Range: `BETWEEN ... AND ...`
- Set membership: `IN (...)`
- Logical: `AND`, `OR`, `NOT`
- Grouping: `(...)`

## Getting an API Key

1. Log in to your Unifi Network Application
2. Navigate to Settings > Control Plane > Integrations
3. Generate a new API key
4. Copy the key and use it in your configuration

## Development

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Clean Build Artifacts

```bash
make clean
```

### Lint

```bash
make lint
```

(Requires [golangci-lint](https://golangci-lint.run/) to be installed)

## Project Structure

```
unifi-cli/
├── cmd/               # Command definitions
│   ├── root.go       # Root command and global flags
│   └── clients.go    # Clients command
├── internal/
│   ├── api/          # API client and types
│   │   ├── client.go
│   │   └── types.go
│   ├── config/       # Configuration management
│   │   └── config.go
│   ├── filter/       # Client filtering with SQLite
│   │   ├── filter.go
│   │   └── schema.go
│   └── output/       # Output formatting
│       ├── table.go
│       └── json.go
├── main.go           # Entry point
├── Makefile          # Build automation
└── README.md         # This file
```

## License

MIT
