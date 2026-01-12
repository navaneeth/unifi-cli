package filter

// clientTableSchema defines the JSON-based SQLite schema
// Store entire client as JSON in single column, with a view for querying
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
