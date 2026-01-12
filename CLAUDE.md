# Claude Project Notes - Unifi CLI

This document contains comprehensive implementation details for future reference and development.

## Project Overview

**Project Name**: Unifi CLI
**Language**: Go (Golang)
**Purpose**: Command-line interface for managing Unifi Network devices using the official Unifi Network API
**Created**: 2026-01-11

## Architecture

### Technology Stack

- **CLI Framework**: [Cobra](https://github.com/spf13/cobra) v1.10.2 - Command structure and parsing
- **Configuration**: [Viper](https://github.com/spf13/viper) v1.21.0 - Config file, env vars, and flags management
- **Table Output**: [tablewriter](https://github.com/olekukonko/tablewriter) v1.1.2 - ASCII table formatting

### Project Structure

```
unifi-cli/
├── cmd/
│   ├── root.go           # Root command with global flags and config initialization
│   └── clients.go        # Clients subcommand (list connected clients)
├── internal/
│   ├── api/
│   │   ├── client.go     # HTTP API client implementation
│   │   ├── client_test.go # API client tests (94.8% coverage)
│   │   ├── types.go      # API response types and helper methods
│   │   └── types_test.go # Type tests
│   ├── config/
│   │   ├── config.go     # Viper-based configuration management
│   │   └── config_test.go # Config tests (93.1% coverage)
│   └── output/
│       ├── table.go      # Table output formatter
│       ├── table_test.go # Table output tests
│       ├── json.go       # JSON output formatter
│       └── json_test.go  # JSON output tests (91.7% coverage)
├── main.go               # Entry point - calls cmd.Execute()
├── Makefile              # Build automation
├── go.mod                # Go module definition
├── go.sum                # Dependency checksums
├── spec.md               # Original specification
├── README.md             # User documentation
├── .gitignore            # Git ignore rules
└── CLAUDE.md             # This file
```

## Implementation Details

### 1. Configuration System

**File**: `internal/config/config.go`

**Configuration Priority** (highest to lowest):
1. Command-line flags
2. Environment variables (with `UNIFI_` prefix)
3. Config file (`~/.unifi-cli.yaml`)
4. Default values

**Environment Variables**:
- `UNIFI_HOST` - Controller host URL
- `UNIFI_API_KEY` - API authentication key
- `UNIFI_SITE` - Site ID (default: "default")

**Default Values**:
- `site`: "default"
- `insecure`: true (TLS verification skipped by default for self-signed certs)

**Important Functions**:
- `Init(cfgFile string)` - Initialize configuration from file or default location
- `Get()` - Returns singleton config instance
- `Validate()` - Validates required fields (host, api_key)
- `GetConfigPath()` - Returns current or default config file path

### 2. API Client

**File**: `internal/api/client.go`

**Key Design Decisions**:
- Uses `APIClient` struct (not `Client` to avoid naming conflict with `Client` type in types.go)
- TLS verification skipped by default (`InsecureSkipVerify: true`)
- 30-second HTTP timeout
- Trailing slash automatically removed from host URL

**API Endpoints**:
- List Clients: `{host}/proxy/network/api/s/{site}/stat/sta`
- List Sites: `{host}/proxy/network/api/self/sites`

**Authentication**:
- Header: `X-API-KEY: {api_key}`
- Content-Type: `application/json`

**Important Note**: The API returns floating-point numbers for `tx_bytes-r` and `rx_bytes-r` fields, so these are typed as `float64` in the `Client` struct.

### 3. Data Types

**File**: `internal/api/types.go`

**Main Types**:
```go
type Client struct {
    MAC       string
    Name      string
    Hostname  string
    IP        string
    IsWired   bool
    Essid     string  // SSID for wireless
    Signal    int     // dBm for wireless
    Uptime    int64   // seconds
    RxBytes   int64
    TxBytes   int64
    TxBytesR  float64 // IMPORTANT: float64 not int64!
    RxBytesR  float64 // IMPORTANT: float64 not int64!
    // ... many more fields
}
```

**Helper Methods**:
- `GetDisplayName()` - Returns Name → Hostname → MAC (fallback chain)
- `GetConnectionType()` - Returns "Wired" or "Wireless"
- `GetSSID()` - Returns SSID for wireless clients, empty for wired
- `GetSignal()` - Returns formatted signal strength (e.g., "-65 dBm")
- `GetUptime()` - Returns human-readable uptime (e.g., "5d 3h 15m")

**Utility Functions**:
- `FormatBytes(bytes int64)` - Converts bytes to human-readable format (KB, MB, GB, TB)

### 4. Output Formatters

**Table Format** (`internal/output/table.go`):
- Uses tablewriter library
- Columns: MAC, Name, IP, Type, SSID, Signal, Uptime, RX/TX
- SSID and Signal only shown for wireless clients
- RX/TX shows formatted bytes (e.g., "35.8 MB / 25.5 MB")

**JSON Format** (`internal/output/json.go`):
- Pretty-printed with 2-space indentation
- Returns raw API response data

### 5. Commands

**Root Command** (`cmd/root.go`):
```bash
unifi [command] [flags]
```

**Global Flags**:
- `--config, -c` - Config file path
- `--host` - Unifi controller host
- `--site` - Site ID (default: "default")
- `--insecure, -k` - Skip TLS verification (default: true)

**Clients Command** (`cmd/clients.go`):
```bash
unifi clients list [flags]
```

**Flags**:
- `--format, -f` - Output format: "table" (default) or "json"

## Build System

**Makefile Targets**:
```makefile
make build      # Build to ./bin/unifi
make compile    # Alias for build
make install    # Run 'go install'
make clean      # Remove ./bin directory
make test       # Run tests
make lint       # Run golangci-lint (requires installation)
make help       # Show available targets
```

## Testing

**Test Coverage**:
- internal/api: 94.8%
- internal/config: 93.1%
- internal/output: 91.7%

**Test Files**:
- `client_test.go` - Uses httptest for mocking API responses
- `config_test.go` - Tests config loading, validation, environment variable handling
- `types_test.go` - Tests all helper methods and formatters
- `json_test.go` - Captures stdout and validates JSON output
- `table_test.go` - Validates table formatting and content

**Running Tests**:
```bash
go test ./...              # Run all tests
go test ./... -v           # Verbose output
go test ./... -cover       # With coverage
```

## Known Issues & Solutions

### Issue 1: Type Mismatch for tx_bytes-r and rx_bytes-r

**Problem**: API returns floating-point numbers but code expected int64
**Error**: `json: cannot unmarshal number 43.08318972270092 into Go struct field Client.data.tx_bytes-r of type int64`
**Solution**: Changed type from `int64` to `float64` in `internal/api/types.go`

### Issue 2: Test Failures with Environment Variables

**Problem**: Config tests failed because UNIFI_API_KEY from environment overrode test config
**Solution**: Added env var cleanup in `TestInitWithValidConfigFile`

### Issue 3: Naming Conflict

**Problem**: Both API client and response data use "Client" name
**Solution**: Renamed API client struct to `APIClient`, kept data type as `Client`

## API Documentation

**Official Unifi API Documentation**:
- Access via: Unifi Network > Settings > Control Plane > Integrations
- Generate API keys in the same location
- Documentation is version-specific to your controller

**API Response Structure**:
```json
{
  "meta": {
    "rc": "ok"  // "ok" for success, "error" for failure
  },
  "data": [
    {
      "_id": "...",
      "mac": "aa:bb:cc:dd:ee:ff",
      "ip": "192.168.1.100",
      "is_wired": true,
      // ... many more fields
    }
  ]
}
```

## Usage Examples

### Basic Usage
```bash
# Set environment variables
export UNIFI_HOST="https://unifi.example.com"
export UNIFI_API_KEY="your-api-key"

# List clients (table format)
./bin/unifi clients list

# List clients (JSON format)
./bin/unifi clients list --format json
./bin/unifi clients list -f json
```

### With Config File
```bash
# Create config file
cat > ~/.unifi-cli.yaml <<EOF
host: https://unifi.example.com
api_key: your-api-key
site: default
insecure: true
EOF

# Use it
./bin/unifi clients list
```

### With Command-line Flags
```bash
./bin/unifi --host https://unifi.example.com clients list
./bin/unifi --site custom-site clients list
```

## Future Extension Points

### Adding New Commands

1. **Create command file** in `cmd/` (e.g., `devices.go`)
2. **Define command structure**:
```go
var devicesCmd = &cobra.Command{
    Use:   "devices",
    Short: "Manage Unifi devices",
}

var devicesListCmd = &cobra.Command{
    Use:   "list",
    Short: "List devices",
    RunE:  runDevicesList,
}

func init() {
    rootCmd.AddCommand(devicesCmd)
    devicesCmd.AddCommand(devicesListCmd)
}
```

3. **Add API method** to `internal/api/client.go`
4. **Add response types** to `internal/api/types.go`
5. **Add output formatters** to `internal/output/`

### Suggested Commands for Future

- `unifi sites list` - List all sites
- `unifi devices list` - List network devices (APs, switches, etc.)
- `unifi networks list` - List networks/SSIDs
- `unifi clients block <mac>` - Block a client
- `unifi clients unblock <mac>` - Unblock a client
- `unifi devices restart <mac>` - Restart a device

### Adding New Output Formats

1. Add formatter to `internal/output/`
2. Update `--format` flag validation
3. Add format to switch statement in command

Example formats to add:
- CSV
- YAML
- Colored table (using fatih/color)
- Wide table (more columns)

## Dependencies

**Direct Dependencies**:
```go
github.com/spf13/cobra v1.10.2
github.com/spf13/viper v1.21.0
github.com/olekukonko/tablewriter v1.1.2
```

**Test Dependencies**:
- Standard library: `testing`, `net/http/httptest`

**Transitive Dependencies** (auto-managed):
- github.com/spf13/pflag
- github.com/spf13/afero
- github.com/spf13/cast
- github.com/fsnotify/fsnotify
- And more (see go.sum)

## Development Workflow

1. **Make changes** to source files
2. **Run tests**: `make test`
3. **Build**: `make build`
4. **Test manually**: `./bin/unifi clients list`
5. **Install**: `make install` (copies to $GOPATH/bin)

## Git Information

**Repository**: Local repository (no remote configured initially)
**Branch**: master
**Ignore patterns**: See `.gitignore`
- bin/ directory
- .unifi-cli.yaml config file
- Standard Go ignores (*.test, *.out)
- IDE files (.vscode, .idea)

## Security Considerations

1. **TLS Verification**: Disabled by default for self-signed certs (common in Unifi setups)
   - Can be enabled by setting `--insecure=false`

2. **API Key Storage**:
   - Never commit API keys to git
   - Config file is in .gitignore
   - Prefer environment variables for CI/CD

3. **Credential Exposure**:
   - API key visible in process list if passed via command line
   - Prefer env vars or config file

## Performance Notes

- HTTP timeout: 30 seconds
- No caching implemented
- Each command makes fresh API call
- Table rendering is fast even with 100+ clients

## Troubleshooting

### "Missing API key or host"
- Ensure `UNIFI_HOST` and `UNIFI_API_KEY` are set
- Or configure in `~/.unifi-cli.yaml`
- Or pass via `--host` flag

### "TLS certificate verification failed"
- Expected with self-signed certs
- Already handled: `--insecure` defaults to true
- To enforce verification: `--insecure=false`

### "API request failed with status 401"
- Invalid API key
- Check key in Unifi controller: Settings > Control Plane > Integrations

### "Cannot unmarshal number ... into type int64"
- This was fixed in the implementation
- If you see it, a field type needs changing from int64 to float64

## Code Quality

- All exported functions have comments
- Error messages include context
- Tests cover edge cases and error paths
- Code follows Go conventions
- No external logging framework (uses standard fmt)

## Lessons Learned

1. **Always check API response types** - The tx_bytes-r float issue
2. **Environment variables can interfere with tests** - Need cleanup
3. **Naming conflicts are common** - Use descriptive names (APIClient vs Client)
4. **Viper automatically reads environment variables** - Set prefix with SetEnvPrefix
5. **Table output formatting is tricky** - tablewriter API changed between versions

## Version Information

- **Go Version**: 1.21+ (uses go.mod)
- **Tested on**: Linux 6.17.12-300.fc43.x86_64
- **Unifi Network Application**: Works with versions supporting the official API (9.1.105+)

## Contact & Support

For Unifi API documentation:
- [Official Unifi API Docs](https://developer.ui.com/site-manager-api/gettingstarted)
- Controller-specific docs: Navigate to Settings > Control Plane > Integrations in your controller

## Changelog

### 2026-01-11 - Initial Implementation
- Created full CLI structure with Cobra and Viper
- Implemented `clients list` command
- Added table and JSON output formats
- Created comprehensive test suite (>90% coverage)
- Fixed float64 type issue for tx_bytes-r and rx_bytes-r
- Added Makefile for build automation
- Documentation: README.md, spec.md, CLAUDE.md
