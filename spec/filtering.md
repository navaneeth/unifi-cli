# Client Filtering Implementation Plan

## Overview
Add filtering capability to `unifi clients list` command using SQLite with SQL WHERE clause syntax. This provides powerful, familiar query capabilities without writing a custom parser.

## Architecture Decision

**SQLite-Based Client-Side Filtering**
- Leverage SQLite's mature query parser instead of custom implementation
- Use **modernc.org/sqlite** - pure Go implementation (no cgo/C dependencies)
- In-memory database (`:memory:`) - no disk I/O, fast for small datasets
- The Unifi API endpoint doesn't support filtering query parameters
- Typical network has 50-200 clients (small dataset, minimal overhead)

## Implementation Design

### 1. New Package: `internal/filter/`

Create a dedicated filter package following the existing architecture pattern:

**Files to create:**
- `internal/filter/filter.go` - SQLite-based filtering engine
- `internal/filter/filter_test.go` - Unit tests (>95% coverage)
- `internal/filter/schema.go` - SQL schema generation from Client struct
- `internal/filter/schema_test.go` - Schema tests

**Dependencies to add:**
```bash
go get modernc.org/sqlite  # Pure Go SQLite implementation
```

### 2. Filter Syntax Design

**SQL WHERE Clause Syntax:**
```bash
# Simple conditions
--filter "signal >= -65"
--filter "is_wired = 1"
--filter "essid = 'HomeWiFi'"

# Complex queries with AND/OR
--filter "signal >= -65 AND essid = 'HomeWiFi'"
--filter "is_wired = 1 OR ap_mac = 'aa:bb:cc:dd:ee:ff'"

# SQL operators
--filter "signal BETWEEN -70 AND -50"
--filter "essid LIKE '%Guest%'"
--filter "hostname IN ('iphone', 'ipad', 'macbook')"
--filter "signal >= -65 AND (essid = 'HomeWiFi' OR essid = 'GuestWiFi')"
```

**Simple Flags (Convenience - Translated to SQL):**
```bash
--wired       # Translates to: is_wired = 1
--wireless    # Translates to: is_wired = 0
--ap <MAC>    # Translates to: ap_mac = '<MAC>'
--blocked     # Translates to: blocked = 1
```

**Field Naming Convention: snake_case**
Go struct field → SQL column mapping:
- `IsWired` → `is_wired`
- `ApMAC` → `ap_mac`
- `Signal` → `signal`
- `Essid` → `essid`
- `Hostname` → `hostname`
- etc.

**Key Filterable Fields:**
- `is_wired`, `blocked`, `use_fixed_ip`, `qos_policy_applied` (bool → integer 0/1)
- `mac`, `name`, `hostname`, `ip`, `essid`, `ap_mac`, `sw_mac` (string → TEXT)
- `signal`, `uptime`, `tx_rate`, `rx_rate`, `channel`, `sw_port` (int → INTEGER)
- `tx_bytes`, `rx_bytes`, `tx_bytes_r`, `rx_bytes_r` (int64/float64 → REAL)

### 3. Implementation Structure

**File: internal/filter/filter.go**

```go
package filter

import (
    "database/sql"
    "encoding/json"
    "fmt"
    _ "modernc.org/sqlite"
    "unifi-cli/internal/api"
)

// Filter applies SQL WHERE clause to clients using JSON storage
type Filter struct {
    db          *sql.DB
    whereClause string
}

// NewFilter creates in-memory SQLite database and returns filter
func NewFilter(whereClause string) (*Filter, error) {
    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // Create table and view
    if _, err := db.Exec(clientTableSchema); err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to create schema: %w", err)
    }

    return &Filter{db: db, whereClause: whereClause}, nil
}

// Apply filters clients using SQL WHERE clause
func (f *Filter) Apply(clients []api.Client) ([]api.Client, error) {
    // Insert clients as JSON
    if err := f.insertClients(clients); err != nil {
        return nil, err
    }

    // Query with WHERE clause
    return f.queryClients()
}

// insertClients inserts all clients as JSON into the database
func (f *Filter) insertClients(clients []api.Client) error {
    tx, err := f.db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare("INSERT INTO clients (data) VALUES (?)")
    if err != nil {
        return fmt.Errorf("failed to prepare statement: %w", err)
    }
    defer stmt.Close()

    for _, client := range clients {
        jsonData, err := json.Marshal(client)
        if err != nil {
            return fmt.Errorf("failed to marshal client: %w", err)
        }

        if _, err := stmt.Exec(string(jsonData)); err != nil {
            return fmt.Errorf("failed to insert client: %w", err)
        }
    }

    return tx.Commit()
}

// queryClients executes SELECT with WHERE clause on the view
func (f *Filter) queryClients() ([]api.Client, error) {
    query := fmt.Sprintf("SELECT data FROM clients_view WHERE %s", f.whereClause)

    rows, err := f.db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to query clients: %w", err)
    }
    defer rows.Close()

    var result []api.Client
    for rows.Next() {
        var jsonData string
        if err := rows.Scan(&jsonData); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }

        var client api.Client
        if err := json.Unmarshal([]byte(jsonData), &client); err != nil {
            return nil, fmt.Errorf("failed to unmarshal client: %w", err)
        }

        result = append(result, client)
    }

    return result, rows.Err()
}

// Close cleans up database connection
func (f *Filter) Close() error {
    return f.db.Close()
}
```

**File: internal/filter/schema.go**

```go
package filter

// JSON-based schema - store entire client as JSON in single column
// Use VIEW to expose filterable fields via json_extract()
const clientTableSchema = `
CREATE TABLE clients (data TEXT);

CREATE VIEW clients_view AS
  SELECT
    data,
    json_extract(data, '$.mac') as mac,
    json_extract(data, '$.name') as name,
    json_extract(data, '$.hostname') as hostname,
    json_extract(data, '$.ip') as ip,
    json_extract(data, '$.is_wired') as is_wired,
    json_extract(data, '$.blocked') as blocked,
    json_extract(data, '$.essid') as essid,
    json_extract(data, '$.ap_mac') as ap_mac,
    json_extract(data, '$.signal') as signal,
    json_extract(data, '$.uptime') as uptime,
    json_extract(data, '$.tx_rate') as tx_rate,
    json_extract(data, '$.rx_rate') as rx_rate,
    json_extract(data, '$.satisfaction') as satisfaction,
    json_extract(data, '$.sw_mac') as sw_mac,
    json_extract(data, '$.sw_port') as sw_port,
    json_extract(data, '$.channel') as channel,
    json_extract(data, '$.rssi') as rssi,
    json_extract(data, '$.tx_bytes') as tx_bytes,
    json_extract(data, '$.rx_bytes') as rx_bytes
  FROM clients;
`

// Benefits:
// - Single JSON column - no complex type mapping
// - View provides clean column names for WHERE clauses
// - Easy to add more filterable fields
// - No reflection or code generation needed
```

### 4. CLI Integration

**File: cmd/clients.go**

Add flags:
```go
var (
    outputFormat   string
    filterWired    bool
    filterWireless bool
    filterBlocked  bool
    filterAP       string
    filterSQL      string
)
```

Add to `init()`:
```go
clientsListCmd.Flags().BoolVar(&filterWired, "wired", false, "Show only wired clients")
clientsListCmd.Flags().BoolVar(&filterWireless, "wireless", false, "Show only wireless clients")
clientsListCmd.Flags().BoolVar(&filterBlocked, "blocked", false, "Show only blocked clients")
clientsListCmd.Flags().StringVar(&filterAP, "ap", "", "Filter by Access Point MAC")
clientsListCmd.Flags().StringVar(&filterSQL, "filter", "", "SQL WHERE clause (e.g., 'signal >= -65 AND essid = \"HomeWiFi\"')")
```

Modify `runClientsList()`:
```go
func runClientsList(cmd *cobra.Command, args []string) error {
    cfg := config.Get()
    apiClient := api.NewAPIClient(cfg.Host, cfg.APIKey, cfg.Site, cfg.Insecure)

    clients, err := apiClient.ListClients()
    if err != nil {
        return fmt.Errorf("failed to list clients: %w", err)
    }

    // Build WHERE clause from flags
    whereClause, err := buildWhereClause()
    if err != nil {
        return err
    }

    // Apply filter if needed
    filteredClients := clients
    if whereClause != "" {
        filterEngine, err := filter.NewFilter(whereClause)
        if err != nil {
            return fmt.Errorf("failed to create filter: %w", err)
        }
        defer filterEngine.Close()

        filteredClients, err = filterEngine.Apply(clients)
        if err != nil {
            return fmt.Errorf("failed to apply filter: %w", err)
        }
    }

    if len(filteredClients) == 0 {
        fmt.Println("No clients match the specified filters")
        return nil
    }

    // Output formatting (unchanged)
    switch outputFormat {
    case "json":
        return output.PrintClientsJSON(filteredClients)
    case "table":
        output.PrintClientsTable(filteredClients)
        return nil
    default:
        return fmt.Errorf("invalid output format: %s", outputFormat)
    }
}

func buildWhereClause() (string, error) {
    var conditions []string

    // Validate mutually exclusive flags
    if filterWired && filterWireless {
        return "", fmt.Errorf("--wired and --wireless are mutually exclusive")
    }

    // Build conditions from simple flags
    if filterWired {
        conditions = append(conditions, "is_wired = 1")
    }
    if filterWireless {
        conditions = append(conditions, "is_wired = 0")
    }
    if filterBlocked {
        conditions = append(conditions, "blocked = 1")
    }
    if filterAP != "" {
        conditions = append(conditions, fmt.Sprintf("ap_mac = '%s'", filterAP))
    }

    // Add custom SQL filter
    if filterSQL != "" {
        conditions = append(conditions, fmt.Sprintf("(%s)", filterSQL))
    }

    if len(conditions) == 0 {
        return "", nil
    }

    return strings.Join(conditions, " AND "), nil
}
```

### 5. Implementation Steps

**Step 1: Add Dependency**
```bash
go get modernc.org/sqlite
```

**Step 2: Create JSON Schema**
- Create `internal/filter/schema.go`
- Define simple JSON-based table: `CREATE TABLE clients (data TEXT)`
- Define view with json_extract for ~20 filterable fields
- No complex type mapping needed - SQLite handles JSON natively

**Step 3: Implement Filter Engine**
- Create `internal/filter/filter.go`
- Implement `NewFilter()` - create in-memory DB and execute schema
- Implement `insertClients()` - marshal to JSON and bulk insert
- Implement `queryClients()` - query view, unmarshal JSON back to []Client
- Implement `Apply()` - orchestrate the flow
- Implement `Close()` - cleanup

**Step 4: Add CLI Flags**
- Update `cmd/clients.go` with new flags
- Implement `buildWhereClause()` helper
- Integrate filter into `runClientsList()`
- Add validation for mutually exclusive flags

**Step 5: Testing**
- Write unit tests for schema generation
- Write unit tests for filter operations
- Write integration tests for CLI
- Test error conditions

### 6. Testing Strategy

**Unit Tests: internal/filter/filter_test.go**

```go
// Database & Schema Tests
- TestNewFilter_CreatesInMemoryDB
- TestNewFilter_CreatesViewSuccessfully
- TestInsertClients_JSONMarshalAndInsert
- TestInsertClients_Transaction

// Query Tests with Mock Data
- TestApply_SimpleCondition (e.g., "signal >= -65")
- TestApply_MultipleAndConditions
- TestApply_OrConditions
- TestApply_BetweenOperator
- TestApply_LikeOperator
- TestApply_InOperator
- TestApply_ComplexWithParentheses
- TestApply_BooleanFields
- TestApply_StringFields
- TestApply_NumericFields
- TestApply_NoMatches_ReturnsEmpty
- TestApply_EmptyWhere_ReturnsAll

// Error Handling Tests
- TestApply_InvalidSQL_ReturnsError
- TestApply_UnknownColumn_ReturnsError
- TestApply_TypeMismatch_ReturnsError
- TestClose_CleansUpDatabase
```

**Integration Tests: cmd/clients_test.go**

```go
- TestClientsListCmd_WithSQLFilter
- TestClientsListCmd_WiredFlag
- TestClientsListCmd_WirelessFlag
- TestClientsListCmd_APFlag
- TestClientsListCmd_CombinedFlags
- TestClientsListCmd_MutuallyExclusiveError
- TestClientsListCmd_InvalidSQL
- TestClientsListCmd_FilterWithJSON
- TestClientsListCmd_FilterWithTable
```

**Test Data Strategy:**
- Create 10-20 mock clients with varied data
- Cover all field types (bool, string, int, float)
- Include edge cases (nulls, empty strings, extreme values)
- Use httptest to mock API responses

**Coverage Target: >95%** for filter package

### 7. Error Handling

**User-Friendly Error Messages:**

```
Invalid SQL syntax in filter: near "INVALID": syntax error
Did you mean: --filter "signal >= -65"?

Unknown column 'invalid_field' in WHERE clause
Available fields: is_wired, ap_mac, signal, essid, hostname, ...

--wired and --wireless flags cannot be used together

Filter returned no matching clients
```

**Error Wrapping:**
- Wrap SQLite errors with context
- Suggest corrections for common mistakes
- List available fields on unknown column errors

### 8. Example Usage

```bash
# Simple flags (convenience)
unifi clients list --wired
unifi clients list --wireless
unifi clients list --ap aa:bb:cc:dd:ee:ff
unifi clients list --blocked

# SQL WHERE clause (powerful)
unifi clients list --filter "signal >= -65"
unifi clients list --filter "essid = 'HomeWiFi'"
unifi clients list --filter "is_wired = 1 AND signal >= -60"
unifi clients list --filter "essid LIKE '%Guest%'"
unifi clients list --filter "signal BETWEEN -70 AND -50"
unifi clients list --filter "hostname IN ('iphone', 'ipad')"
unifi clients list --filter "(essid = 'HomeWiFi' OR essid = 'GuestWiFi') AND signal >= -65"

# Combined (flags + SQL)
unifi clients list --wireless --filter "signal >= -65"
unifi clients list --ap aa:bb:cc:dd:ee:ff --filter "hostname LIKE '%phone%'"

# With output formats
unifi clients list --filter "signal < -75" --format json
unifi clients list --wired --format table
```

### 9. Performance Considerations

**Expected Performance:**
- DB creation: ~1-2ms
- Schema creation: ~1ms
- Bulk insert (200 clients): ~5-10ms
- Query execution: ~1-5ms
- **Total overhead: ~10-20ms** for typical network

**Optimization Notes:**
- In-memory database eliminates disk I/O
- Single transaction for bulk insert
- Prepared statements for queries
- Connection pooling not needed (single query per command)

### 10. Verification & Testing

**Manual Testing Checklist:**

```bash
# 1. Build
make build

# 2. Test simple flags
./bin/unifi clients list --wired
./bin/unifi clients list --wireless
./bin/unifi clients list --ap <real-ap-mac>

# 3. Test SQL filters
./bin/unifi clients list --filter "signal >= -65"
./bin/unifi clients list --filter "essid = 'HomeWiFi'"
./bin/unifi clients list --filter "is_wired = 1 AND ap_mac = 'aa:bb:cc:dd:ee:ff'"
./bin/unifi clients list --filter "essid LIKE '%Guest%'"

# 4. Test combined
./bin/unifi clients list --wireless --filter "signal >= -70"

# 5. Test error conditions
./bin/unifi clients list --wired --wireless  # Should error
./bin/unifi clients list --filter "invalid_field = 123"  # Should error
./bin/unifi clients list --filter "signal = 'abc'"  # Type error
./bin/unifi clients list --filter "INVALID SQL HERE"  # Syntax error

# 6. Test output formats
./bin/unifi clients list --wireless --format json
./bin/unifi clients list --filter "blocked = 1" --format table

# 7. Run unit tests
make test

# 8. Check coverage
go test ./internal/filter -cover
```

**Verification Criteria:**
- ✅ All unit tests pass with >95% coverage
- ✅ Integration tests pass
- ✅ Manual testing with real Unifi controller works
- ✅ Error messages are clear and helpful
- ✅ Performance overhead <20ms for typical networks
- ✅ Documentation updated (README.md, CLAUDE.md)

### 11. Critical Files

**To Create:**
- `/home/nkn/code/unifi-cli/internal/filter/filter.go` - Core filtering engine
- `/home/nkn/code/unifi-cli/internal/filter/filter_test.go` - Filter tests
- `/home/nkn/code/unifi-cli/internal/filter/schema.go` - SQL schema definition
- `/home/nkn/code/unifi-cli/internal/filter/schema_test.go` - Schema tests

**To Modify:**
- `/home/nkn/code/unifi-cli/cmd/clients.go` - Add flags and integrate filter
- `/home/nkn/code/unifi-cli/go.mod` - Add modernc.org/sqlite dependency
- `/home/nkn/code/unifi-cli/README.md` - Document filtering usage
- `/home/nkn/code/unifi-cli/CLAUDE.md` - Update with implementation details

---

## Summary

This SQLite JSON-based approach provides:
- ✅ **Powerful SQL WHERE clause syntax** - No custom parser needed
- ✅ **Pure Go implementation** - No cgo/C dependencies (modernc.org/sqlite)
- ✅ **Familiar syntax** - SQL is widely known
- ✅ **Minimal overhead** - ~10-20ms for in-memory operations
- ✅ **Simple implementation** - JSON storage eliminates complex type mapping and reflection
- ✅ **Single column storage** - Store entire Client as JSON, query via view
- ✅ **Extensible** - Easy to add new filterable fields (just add to view)
- ✅ **Hybrid approach** - Simple flags for common cases, SQL for complex queries
- ✅ **Well-tested** - Comprehensive unit and integration tests (>95% coverage)
- ✅ **Automatic type handling** - SQLite JSON functions handle all type conversions
