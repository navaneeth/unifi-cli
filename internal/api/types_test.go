package api

import (
	"testing"
	"time"
)

func TestClient_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		client   Client
		expected string
	}{
		{
			name: "name is set",
			client: Client{
				Name:     "MyDevice",
				Hostname: "hostname",
				OUI:      "Sony Interactive Entertainment Inc.",
				MAC:      "aa:bb:cc:dd:ee:ff",
			},
			expected: "MyDevice",
		},
		{
			name: "only hostname is set",
			client: Client{
				Name:     "",
				Hostname: "hostname",
				OUI:      "Sony Interactive Entertainment Inc.",
				MAC:      "aa:bb:cc:dd:ee:ff",
			},
			expected: "hostname",
		},
		{
			name: "only OUI is set",
			client: Client{
				Name:     "",
				Hostname: "",
				OUI:      "Sony Interactive Entertainment Inc.",
				MAC:      "aa:bb:cc:dd:ee:ff",
			},
			expected: "Sony Interactive Entertainment Inc.",
		},
		{
			name: "only MAC is set",
			client: Client{
				Name:     "",
				Hostname: "",
				OUI:      "",
				MAC:      "aa:bb:cc:dd:ee:ff",
			},
			expected: "aa:bb:cc:dd:ee:ff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.client.GetDisplayName()
			if result != tt.expected {
				t.Errorf("GetDisplayName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClient_GetConnectionType(t *testing.T) {
	tests := []struct {
		name     string
		client   Client
		expected string
	}{
		{
			name: "wired connection",
			client: Client{
				IsWired: true,
			},
			expected: "Wired",
		},
		{
			name: "wireless connection",
			client: Client{
				IsWired: false,
			},
			expected: "Wireless",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.client.GetConnectionType()
			if result != tt.expected {
				t.Errorf("GetConnectionType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClient_GetSSID(t *testing.T) {
	tests := []struct {
		name     string
		client   Client
		expected string
	}{
		{
			name: "wireless with SSID",
			client: Client{
				IsWired: false,
				Essid:   "MyWiFi",
			},
			expected: "MyWiFi",
		},
		{
			name: "wired connection",
			client: Client{
				IsWired: true,
				Essid:   "MyWiFi",
			},
			expected: "",
		},
		{
			name: "wireless without SSID",
			client: Client{
				IsWired: false,
				Essid:   "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.client.GetSSID()
			if result != tt.expected {
				t.Errorf("GetSSID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClient_GetSignal(t *testing.T) {
	tests := []struct {
		name     string
		client   Client
		expected string
	}{
		{
			name: "wireless with signal",
			client: Client{
				IsWired: false,
				Signal:  -65,
			},
			expected: "-65 dBm",
		},
		{
			name: "wired connection",
			client: Client{
				IsWired: true,
				Signal:  -65,
			},
			expected: "",
		},
		{
			name: "wireless with zero signal",
			client: Client{
				IsWired: false,
				Signal:  0,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.client.GetSignal()
			if result != tt.expected {
				t.Errorf("GetSignal() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClient_GetUptime(t *testing.T) {
	tests := []struct {
		name     string
		uptime   int64
		expected string
	}{
		{
			name:     "less than one hour",
			uptime:   2700, // 45 minutes
			expected: "45m",
		},
		{
			name:     "hours and minutes",
			uptime:   7500, // 2 hours 5 minutes
			expected: "2h 5m",
		},
		{
			name:     "days hours and minutes",
			uptime:   356700, // 4 days 3 hours 5 minutes
			expected: "4d 3h 5m",
		},
		{
			name:     "exactly one day",
			uptime:   86400, // 1 day
			expected: "1d",
		},
		{
			name:     "zero uptime",
			uptime:   0,
			expected: "0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{Uptime: tt.uptime}
			result := client.GetUptime()
			if result != tt.expected {
				t.Errorf("GetUptime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "less than 1 KB",
			bytes:    512,
			expected: "512 B",
		},
		{
			name:     "KB range",
			bytes:    2048, // 2 KB
			expected: "2.00 KB",
		},
		{
			name:     "MB range",
			bytes:    5242880, // 5 MB
			expected: "5.00 MB",
		},
		{
			name:     "GB range",
			bytes:    3221225472, // 3 GB
			expected: "3.00 GB",
		},
		{
			name:     "TB range",
			bytes:    1099511627776, // 1 TB
			expected: "1.00 TB",
		},
		{
			name:     "zero bytes",
			bytes:    0,
			expected: "0 B",
		},
		{
			name:     "fractional KB",
			bytes:    1536, // 1.5 KB
			expected: "1.50 KB",
		},
		{
			name:     "large MB value",
			bytes:    52428800, // 50 MB
			expected: "50.0 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %v, want %v", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		unit     string
		expected string
	}{
		{
			name:     "days",
			value:    5,
			unit:     "d",
			expected: "5d",
		},
		{
			name:     "hours",
			value:    12,
			unit:     "h",
			expected: "12h",
		},
		{
			name:     "minutes",
			value:    30,
			unit:     "m",
			expected: "30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.value, tt.unit)
			if result != tt.expected {
				t.Errorf("formatValue(%d, %s) = %v, want %v", tt.value, tt.unit, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		days     time.Duration
		hours    time.Duration
		minutes  time.Duration
		showDays bool
		expected string
	}{
		{
			name:     "with days",
			days:     time.Duration(5),
			hours:    time.Duration(3),
			minutes:  time.Duration(15),
			showDays: true,
			expected: "5d 3h 15m",
		},
		{
			name:     "hours only",
			days:     time.Duration(0),
			hours:    time.Duration(2),
			minutes:  time.Duration(30),
			showDays: false,
			expected: "2h 30m",
		},
		{
			name:     "minutes only",
			days:     time.Duration(0),
			hours:    time.Duration(0),
			minutes:  time.Duration(45),
			showDays: false,
			expected: "45m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(
				tt.days,
				tt.hours,
				tt.minutes,
				tt.showDays,
			)
			if result != tt.expected {
				t.Errorf("formatDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}
