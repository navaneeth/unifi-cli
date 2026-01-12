# Unifi CLI - Specification

## Overview
A command-line interface tool for managing Unifi Network devices using the official Unifi Network API.

## Technical Stack
- **Language**: Go (Golang)
- **CLI Framework**: Cobra (command structure)
- **Configuration**: Viper (env vars, config files)
- **Build System**: Makefile

## Authentication & Configuration
- **API Key Header**: `X-API-KEY`
- **Environment Variables**:
  - `UNIFI_API_KEY`: API key for authentication
  - `UNIFI_HOST`: Unifi controller host (e.g., `https://unifi.example.com`)
  - `UNIFI_SITE`: Site ID (optional, defaults to "default" or auto-detect)

## Global Flags
- `--config, -c`: Path to config file (e.g., `~/.unifi-cli.yaml`)
- `--insecure, -k`: Skip TLS certificate verification (default: true, set to false to verify certificates)
- `--host`: Override Unifi controller host
- `--site`: Override site ID

## Commands

### 1. List Clients
```bash
unifi clients list [flags]
```

**Functionality**:
- Fetches active connected clients
- API Endpoint: `/api/s/{site}/stat/sta`
- Auto-detects site ID if only one site exists

**Flags**:
- `--format, -f`: Output format (`table` (default) or `json`)

**Table Columns** (sensible defaults):
- MAC Address
- Name/Hostname
- IP Address
- Type (Wired/Wireless)
- SSID (wireless only)
- Signal (wireless only, dBm)
- Uptime
- RX/TX (data usage)

## Project Structure
```
unifi-cli/
├── cmd/
│   ├── root.go           # Root command + global flags
│   └── clients.go        # Clients commands
├── internal/
│   ├── api/
│   │   ├── client.go     # API client implementation
│   │   └── types.go      # API response types
│   ├── config/
│   │   └── config.go     # Viper configuration
│   └── output/
│       ├── table.go      # Table formatting
│       └── json.go       # JSON formatting
├── main.go
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## Makefile Targets
```makefile
build:      # Build binary to ./bin/unifi
compile:    # Alias for build
install:    # Run 'go install' to install to $GOPATH/bin
clean:      # Remove build artifacts
test:       # Run tests
help:       # Show available targets
```

## Configuration File Format
```yaml
# ~/.unifi-cli.yaml
host: https://unifi.example.com
api_key: your-api-key-here
site: default
insecure: true  # Skip TLS verification by default (common for self-signed certs)
```

## Configuration Priority (highest to lowest)
1. Command-line flags
2. Environment variables
3. Config file
4. Defaults

## Key Features
- HTTP client with TLS verification skipped by default (common for self-signed certs)
- Site ID auto-detection
- Clean table output with column formatting
- JSON output option
- Proper error handling with clear messages

## API Details
- **Base URL**: `{host}/proxy/network`
- **Auth Header**: `X-API-KEY: {api_key}`
- **Clients Endpoint**: `/api/s/{site}/stat/sta` (active clients)
- **Response Format**: JSON

## Error Handling
- Clear error messages for:
  - Missing API key or host
  - Connection failures
  - Invalid site ID
  - API errors (with HTTP status codes)
- Exit codes: 0 (success), 1 (error)

## Future Extensibility
The CLI structure should allow easy addition of:
- `unifi devices list` - List network devices
- `unifi sites list` - List all sites
- `unifi networks list` - List networks/SSIDs
- Additional client management commands
