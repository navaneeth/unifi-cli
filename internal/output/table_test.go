package output

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/nkn/unifi-cli/internal/api"
)

func TestPrintClientsTable(t *testing.T) {
	tests := []struct {
		name    string
		clients []api.Client
	}{
		{
			name: "single wired client",
			clients: []api.Client{
				{
					MAC:      "aa:bb:cc:dd:ee:ff",
					Name:     "TestDevice",
					IP:       "192.168.1.100",
					IsWired:  true,
					Hostname: "test-host",
					Uptime:   3600,
					RxBytes:  1024000,
					TxBytes:  2048000,
				},
			},
		},
		{
			name: "single wireless client",
			clients: []api.Client{
				{
					MAC:      "11:22:33:44:55:66",
					Name:     "WirelessDevice",
					IP:       "192.168.1.101",
					IsWired:  false,
					Essid:    "MyWiFi",
					Signal:   -65,
					Hostname: "wireless-host",
					Uptime:   7200,
					RxBytes:  5242880,
					TxBytes:  10485760,
				},
			},
		},
		{
			name: "multiple clients",
			clients: []api.Client{
				{
					MAC:     "aa:bb:cc:dd:ee:ff",
					Name:    "Device1",
					IP:      "192.168.1.100",
					IsWired: true,
					Uptime:  3600,
				},
				{
					MAC:     "11:22:33:44:55:66",
					Name:    "Device2",
					IP:      "192.168.1.101",
					IsWired: false,
					Essid:   "TestWiFi",
					Signal:  -70,
					Uptime:  7200,
				},
			},
		},
		{
			name:    "empty client list",
			clients: []api.Client{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// This should not panic
			PrintClientsTable(tt.clients)

			w.Close()
			os.Stdout = oldStdout

			// Read output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Verify output contains expected data
			if len(tt.clients) > 0 {
				// Check for MAC addresses
				for _, client := range tt.clients {
					if !strings.Contains(output, client.MAC) {
						t.Errorf("Output should contain MAC address '%s'", client.MAC)
					}
					if !strings.Contains(output, client.IP) {
						t.Errorf("Output should contain IP address '%s'", client.IP)
					}
				}
			}
		})
	}
}

func TestPrintClientsTable_OutputFormat(t *testing.T) {
	clients := []api.Client{
		{
			MAC:      "aa:bb:cc:dd:ee:ff",
			Name:     "TestDevice",
			IP:       "192.168.1.100",
			IsWired:  true,
			Hostname: "test-host",
			Uptime:   86400, // 1 day
			RxBytes:  1048576,
			TxBytes:  2097152,
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintClientsTable(clients)

	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify table contains expected headers and values
	expectedValues := []string{
		"Name",
		"IP",
		"Type",
		"aa:bb:cc:dd:ee:ff", // MAC should be in output (in parentheses with name)
		"TestDevice",
		"192.168.1.100",
		"Wired",
	}

	for _, expected := range expectedValues {
		if !strings.Contains(output, expected) {
			t.Errorf("Table output should contain '%s'", expected)
		}
	}
}

func TestPrintClientsTable_WiredVsWireless(t *testing.T) {
	clients := []api.Client{
		{
			MAC:     "aa:bb:cc:dd:ee:ff",
			IsWired: true,
			IP:      "192.168.1.1",
		},
		{
			MAC:     "11:22:33:44:55:66",
			IsWired: false,
			Essid:   "TestSSID",
			Signal:  -60,
			IP:      "192.168.1.2",
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintClientsTable(clients)

	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify wired shows "Wired"
	if !strings.Contains(output, "Wired") {
		t.Error("Output should contain 'Wired' for wired client")
	}

	// Verify wireless shows "Wireless"
	if !strings.Contains(output, "Wireless") {
		t.Error("Output should contain 'Wireless' for wireless client")
	}

	// Verify wireless shows SSID
	if !strings.Contains(output, "TestSSID") {
		t.Error("Output should contain SSID for wireless client")
	}

	// Verify wireless shows signal
	if !strings.Contains(output, "dBm") {
		t.Error("Output should contain signal strength for wireless client")
	}
}
