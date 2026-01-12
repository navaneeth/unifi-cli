package api

import (
	"fmt"
	"time"
)

type APIResponse struct {
	Meta Meta          `json:"meta"`
	Data []interface{} `json:"data"`
}

type Meta struct {
	RC string `json:"rc"`
}

type ClientsResponse struct {
	Meta Meta     `json:"meta"`
	Data []Client `json:"data"`
}

type Client struct {
	ID               string  `json:"_id"`
	MAC              string  `json:"mac"`
	SiteID           string  `json:"site_id"`
	AssocTime        int64   `json:"assoc_time"`
	LatestAssocTime  int64   `json:"latest_assoc_time"`
	OUI              string  `json:"oui"`
	UserID           string  `json:"user_id"`
	Uptime           int64   `json:"uptime"`
	LastSeen         int64   `json:"last_seen"`
	IsWired          bool    `json:"is_wired"`
	Hostname         string  `json:"hostname"`
	Name             string  `json:"name"`
	IP               string  `json:"ip"`
	Essid            string  `json:"essid"`
	BSSID            string  `json:"bssid"`
	Channel          int     `json:"channel"`
	Radio            string  `json:"radio"`
	RadioName        string  `json:"radio_name"`
	RadioProto       string  `json:"radio_proto"`
	RSSI             int     `json:"rssi"`
	Signal           int     `json:"signal"`
	Noise            int     `json:"noise"`
	TxRate           int     `json:"tx_rate"`
	RxRate           int     `json:"rx_rate"`
	TxBytes          int64   `json:"tx_bytes"`
	RxBytes          int64   `json:"rx_bytes"`
	TxPackets        int64   `json:"tx_packets"`
	RxPackets        int64   `json:"rx_packets"`
	TxBytesR         float64 `json:"tx_bytes-r"`
	RxBytesR         float64 `json:"rx_bytes-r"`
	Satisfaction     int     `json:"satisfaction"`
	Note             string  `json:"note"`
	ApMAC            string  `json:"ap_mac"`
	SWMAC            string  `json:"sw_mac"`
	SWPort           int     `json:"sw_port"`
	Network          string  `json:"network"`
	NetworkID        string  `json:"network_id"`
	UseFixedIP       bool    `json:"use_fixedip"`
	FixedIP          string  `json:"fixed_ip"`
	DeviceIDOverride int     `json:"deviceIdOverride"`
	Blocked          bool    `json:"blocked"`
	QOSPolicyApplied bool    `json:"qos_policy_applied"`
}

// GetDisplayName returns the best available name for the client
// Fallback order: Name -> Hostname -> OUI (manufacturer) -> MAC
func (c *Client) GetDisplayName() string {
	if c.Name != "" {
		return c.Name
	}
	if c.Hostname != "" {
		return c.Hostname
	}
	if c.OUI != "" {
		return c.OUI
	}
	return c.MAC
}

// GetConnectionType returns "Wired" or "Wireless"
func (c *Client) GetConnectionType() string {
	if c.IsWired {
		return "Wired"
	}
	return "Wireless"
}

// GetSSID returns the SSID for wireless clients, empty for wired
func (c *Client) GetSSID() string {
	if !c.IsWired {
		return c.Essid
	}
	return ""
}

// GetSignal returns the signal strength for wireless clients
func (c *Client) GetSignal() string {
	if !c.IsWired && c.Signal != 0 {
		return fmt.Sprintf("%d dBm", c.Signal)
	}
	return ""
}

// GetUptime returns a human-readable uptime duration
func (c *Client) GetUptime() string {
	d := time.Duration(c.Uptime) * time.Second

	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour

	hours := d / time.Hour
	d -= hours * time.Hour

	minutes := d / time.Minute

	if days > 0 {
		return formatDuration(days, hours, minutes, true)
	}
	if hours > 0 {
		return formatDuration(0, hours, minutes, false)
	}
	return formatDuration(0, 0, minutes, false)
}

func formatDuration(days, hours, minutes time.Duration, showDays bool) string {
	if showDays {
		return formatTime(int(days), int(hours), int(minutes), "d", "h", "m")
	}
	if hours > 0 {
		return formatTime(int(hours), int(minutes), 0, "h", "m", "")
	}
	return formatTime(int(minutes), 0, 0, "m", "", "")
}

func formatTime(v1, v2, v3 int, u1, u2, u3 string) string {
	result := ""
	if v1 > 0 {
		result += formatValue(v1, u1)
	}
	if v2 > 0 {
		if result != "" {
			result += " "
		}
		result += formatValue(v2, u2)
	}
	if v3 > 0 && u3 != "" {
		if result != "" {
			result += " "
		}
		result += formatValue(v3, u3)
	}
	if result == "" {
		return "0m"
	}
	return result
}

func formatValue(v int, unit string) string {
	return fmt.Sprintf("%d%s", v, unit)
}

// FormatBytes returns human-readable bytes
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	value := float64(bytes) / float64(div)

	if value >= 10 {
		return fmt.Sprintf("%.1f %s", value, units[exp])
	}
	return fmt.Sprintf("%.2f %s", value, units[exp])
}
