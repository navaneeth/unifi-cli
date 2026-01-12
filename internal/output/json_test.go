package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/nkn/unifi-cli/internal/api"
)

func TestPrintClientsJSON(t *testing.T) {
	tests := []struct {
		name    string
		clients []api.Client
		wantErr bool
	}{
		{
			name: "single client",
			clients: []api.Client{
				{
					MAC:      "aa:bb:cc:dd:ee:ff",
					Name:     "TestDevice",
					IP:       "192.168.1.100",
					IsWired:  true,
					Hostname: "test-host",
					Uptime:   3600,
				},
			},
			wantErr: false,
		},
		{
			name: "multiple clients",
			clients: []api.Client{
				{
					MAC:     "aa:bb:cc:dd:ee:ff",
					Name:    "Device1",
					IsWired: true,
				},
				{
					MAC:     "11:22:33:44:55:66",
					Name:    "Device2",
					IsWired: false,
					Essid:   "MyWiFi",
					Signal:  -65,
				},
			},
			wantErr: false,
		},
		{
			name:    "empty client list",
			clients: []api.Client{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := PrintClientsJSON(tt.clients)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			if (err != nil) != tt.wantErr {
				t.Errorf("PrintClientsJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Verify it's valid JSON
			var result []api.Client
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Errorf("Output is not valid JSON: %v", err)
				return
			}

			// Verify the data matches
			if len(result) != len(tt.clients) {
				t.Errorf("Expected %d clients in output, got %d", len(tt.clients), len(result))
			}
		})
	}
}

func TestPrintClientsJSON_ValidFormat(t *testing.T) {
	clients := []api.Client{
		{
			MAC:     "aa:bb:cc:dd:ee:ff",
			Name:    "TestDevice",
			IP:      "192.168.1.100",
			IsWired: true,
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := PrintClientsJSON(clients)
	if err != nil {
		t.Fatalf("PrintClientsJSON() returned error: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Parse JSON
	var result []api.Client
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify content
	if len(result) != 1 {
		t.Fatalf("Expected 1 client, got %d", len(result))
	}

	if result[0].MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("Expected MAC 'aa:bb:cc:dd:ee:ff', got '%s'", result[0].MAC)
	}

	if result[0].Name != "TestDevice" {
		t.Errorf("Expected Name 'TestDevice', got '%s'", result[0].Name)
	}
}
