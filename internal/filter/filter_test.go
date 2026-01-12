package filter

import (
	"testing"

	"github.com/nkn/unifi-cli/internal/api"
)

// Helper function to create test clients
func createTestClients() []api.Client {
	return []api.Client{
		{
			MAC:      "aa:bb:cc:dd:ee:01",
			Name:     "iPhone",
			Hostname: "iphone-12",
			IP:       "192.168.1.100",
			IsWired:  false,
			Blocked:  false,
			Essid:    "HomeWiFi",
			ApMAC:    "11:22:33:44:55:66",
			Signal:   -45,
			Uptime:   3600,
		},
		{
			MAC:      "aa:bb:cc:dd:ee:02",
			Name:     "MacBook",
			Hostname: "macbook-pro",
			IP:       "192.168.1.101",
			IsWired:  true,
			Blocked:  false,
			Essid:    "",
			ApMAC:    "",
			Signal:   0,
			Uptime:   7200,
		},
		{
			MAC:      "aa:bb:cc:dd:ee:03",
			Name:     "iPad",
			Hostname: "ipad-air",
			IP:       "192.168.1.102",
			IsWired:  false,
			Blocked:  false,
			Essid:    "GuestWiFi",
			ApMAC:    "11:22:33:44:55:77",
			Signal:   -70,
			Uptime:   1800,
		},
		{
			MAC:      "aa:bb:cc:dd:ee:04",
			Name:     "Desktop",
			Hostname: "desktop-pc",
			IP:       "192.168.1.103",
			IsWired:  true,
			Blocked:  true,
			Essid:    "",
			ApMAC:    "",
			Signal:   0,
			Uptime:   86400,
		},
		{
			MAC:      "aa:bb:cc:dd:ee:05",
			Name:     "Android",
			Hostname: "android-phone",
			IP:       "192.168.1.104",
			IsWired:  false,
			Blocked:  false,
			Essid:    "HomeWiFi",
			ApMAC:    "11:22:33:44:55:66",
			Signal:   -55,
			Uptime:   5400,
		},
	}
}

func TestNewFilter_CreatesInMemoryDB(t *testing.T) {
	f, err := NewFilter("1 = 1")
	if err != nil {
		t.Fatalf("NewFilter failed: %v", err)
	}
	defer f.Close()

	if f.db == nil {
		t.Error("Expected database to be initialized")
	}
	if f.whereClause != "1 = 1" {
		t.Errorf("Expected whereClause '1 = 1', got '%s'", f.whereClause)
	}
}

func TestApply_SimpleCondition(t *testing.T) {
	clients := createTestClients()
	// Filter for wireless clients with good signal (exclude wired clients with signal = 0)
	f, err := NewFilter("signal >= -60 AND signal < 0")
	if err != nil {
		t.Fatalf("NewFilter failed: %v", err)
	}
	defer f.Close()

	result, err := f.Apply(clients)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should match iPhone (-45) and Android (-55)
	if len(result) != 2 {
		t.Errorf("Expected 2 clients with signal >= -60 and signal < 0, got %d", len(result))
	}

	for _, client := range result {
		if client.Signal < -60 || client.Signal >= 0 {
			t.Errorf("Client %s has signal %d, expected >= -60 and < 0", client.MAC, client.Signal)
		}
	}
}

func TestApply_BooleanFields(t *testing.T) {
	clients := createTestClients()

	tests := []struct {
		name     string
		where    string
		expected int
	}{
		{"Wired clients", "is_wired = 1", 2},
		{"Wireless clients", "is_wired = 0", 3},
		{"Blocked clients", "blocked = 1", 1},
		{"Not blocked", "blocked = 0", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFilter(tt.where)
			if err != nil {
				t.Fatalf("NewFilter failed: %v", err)
			}
			defer f.Close()

			result, err := f.Apply(clients)
			if err != nil {
				t.Fatalf("Apply failed: %v", err)
			}

			if len(result) != tt.expected {
				t.Errorf("Expected %d clients, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestApply_StringFields(t *testing.T) {
	clients := createTestClients()

	tests := []struct {
		name     string
		where    string
		expected int
	}{
		{"Exact SSID match", "essid = 'HomeWiFi'", 2},
		{"LIKE operator", "hostname LIKE '%iphone%'", 1},
		{"AP MAC match", "ap_mac = '11:22:33:44:55:66'", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFilter(tt.where)
			if err != nil {
				t.Fatalf("NewFilter failed: %v", err)
			}
			defer f.Close()

			result, err := f.Apply(clients)
			if err != nil {
				t.Fatalf("Apply failed: %v", err)
			}

			if len(result) != tt.expected {
				t.Errorf("Expected %d clients, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestApply_MultipleAndConditions(t *testing.T) {
	clients := createTestClients()
	f, err := NewFilter("is_wired = 0 AND signal >= -60")
	if err != nil {
		t.Fatalf("NewFilter failed: %v", err)
	}
	defer f.Close()

	result, err := f.Apply(clients)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should match iPhone (-45) and Android (-55) - both wireless with good signal
	if len(result) != 2 {
		t.Errorf("Expected 2 wireless clients with signal >= -60, got %d", len(result))
	}

	for _, client := range result {
		if client.IsWired {
			t.Errorf("Client %s is wired, expected wireless", client.MAC)
		}
		if client.Signal < -60 {
			t.Errorf("Client %s has signal %d, expected >= -60", client.MAC, client.Signal)
		}
	}
}

func TestApply_OrConditions(t *testing.T) {
	clients := createTestClients()
	f, err := NewFilter("essid = 'HomeWiFi' OR essid = 'GuestWiFi'")
	if err != nil {
		t.Fatalf("NewFilter failed: %v", err)
	}
	defer f.Close()

	result, err := f.Apply(clients)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should match all wireless clients (3 total)
	if len(result) != 3 {
		t.Errorf("Expected 3 wireless clients, got %d", len(result))
	}
}

func TestApply_ComplexWithParentheses(t *testing.T) {
	clients := createTestClients()
	f, err := NewFilter("(essid = 'HomeWiFi' OR essid = 'GuestWiFi') AND signal >= -60")
	if err != nil {
		t.Fatalf("NewFilter failed: %v", err)
	}
	defer f.Close()

	result, err := f.Apply(clients)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should match iPhone (-45) and Android (-55)
	if len(result) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(result))
	}
}

func TestApply_BetweenOperator(t *testing.T) {
	clients := createTestClients()
	f, err := NewFilter("signal BETWEEN -60 AND -50")
	if err != nil {
		t.Fatalf("NewFilter failed: %v", err)
	}
	defer f.Close()

	result, err := f.Apply(clients)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should match Android (-55)
	if len(result) != 1 {
		t.Errorf("Expected 1 client with signal between -60 and -50, got %d", len(result))
	}
}

func TestApply_InOperator(t *testing.T) {
	clients := createTestClients()
	f, err := NewFilter("hostname IN ('iphone-12', 'macbook-pro')")
	if err != nil {
		t.Fatalf("NewFilter failed: %v", err)
	}
	defer f.Close()

	result, err := f.Apply(clients)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should match iPhone and MacBook
	if len(result) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(result))
	}
}

func TestApply_NoMatches_ReturnsEmpty(t *testing.T) {
	clients := createTestClients()
	f, err := NewFilter("signal < -100")
	if err != nil {
		t.Fatalf("NewFilter failed: %v", err)
	}
	defer f.Close()

	result, err := f.Apply(clients)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 clients, got %d", len(result))
	}
}

func TestApply_AllMatch(t *testing.T) {
	clients := createTestClients()
	f, err := NewFilter("1 = 1")
	if err != nil {
		t.Fatalf("NewFilter failed: %v", err)
	}
	defer f.Close()

	result, err := f.Apply(clients)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if len(result) != len(clients) {
		t.Errorf("Expected %d clients, got %d", len(clients), len(result))
	}
}

func TestApply_InvalidSQL_ReturnsError(t *testing.T) {
	clients := createTestClients()
	f, err := NewFilter("INVALID SQL HERE")
	if err != nil {
		t.Fatalf("NewFilter failed: %v", err)
	}
	defer f.Close()

	_, err = f.Apply(clients)
	if err == nil {
		t.Error("Expected error for invalid SQL, got nil")
	}
}

func TestApply_UnknownColumn_ReturnsError(t *testing.T) {
	clients := createTestClients()
	f, err := NewFilter("invalid_field = 123")
	if err != nil {
		t.Fatalf("NewFilter failed: %v", err)
	}
	defer f.Close()

	_, err = f.Apply(clients)
	if err == nil {
		t.Error("Expected error for unknown column, got nil")
	}
}

func TestClose_CleansUpDatabase(t *testing.T) {
	f, err := NewFilter("1 = 1")
	if err != nil {
		t.Fatalf("NewFilter failed: %v", err)
	}

	err = f.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Try to use after close - should fail
	clients := createTestClients()
	_, err = f.Apply(clients)
	if err == nil {
		t.Error("Expected error when using closed filter, got nil")
	}
}

func TestApply_EmptyClientList(t *testing.T) {
	f, err := NewFilter("1 = 1")
	if err != nil {
		t.Fatalf("NewFilter failed: %v", err)
	}
	defer f.Close()

	result, err := f.Apply([]api.Client{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 clients for empty input, got %d", len(result))
	}
}
