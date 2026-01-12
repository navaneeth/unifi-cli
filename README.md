# Unifi CLI

A command-line interface for managing Unifi Network devices using the official Unifi Network API.

## Features

- List connected clients with detailed information
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
│   └── output/       # Output formatting
│       ├── table.go
│       └── json.go
├── main.go           # Entry point
├── Makefile          # Build automation
└── README.md         # This file
```

## License

MIT
