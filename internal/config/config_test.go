package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestInit(t *testing.T) {
	// Reset viper before tests
	viper.Reset()

	tests := []struct {
		name      string
		cfgFile   string
		wantError bool
	}{
		{
			name:      "no config file specified",
			cfgFile:   "",
			wantError: false,
		},
		{
			name:      "non-existent config file",
			cfgFile:   "/tmp/non-existent-config.yaml",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			err := Init(tt.cfgFile)
			if (err != nil) != tt.wantError {
				t.Errorf("Init() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestInitWithValidConfigFile(t *testing.T) {
	viper.Reset()
	cfg = nil

	// Clear environment variables that might interfere
	oldAPIKey := os.Getenv("UNIFI_API_KEY")
	oldHost := os.Getenv("UNIFI_HOST")
	os.Unsetenv("UNIFI_API_KEY")
	os.Unsetenv("UNIFI_HOST")
	defer func() {
		if oldAPIKey != "" {
			os.Setenv("UNIFI_API_KEY", oldAPIKey)
		}
		if oldHost != "" {
			os.Setenv("UNIFI_HOST", oldHost)
		}
	}()

	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `host: https://test.example.com
api_key: test-api-key
site: test-site
insecure: false
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	if err := Init(configFile); err != nil {
		t.Fatalf("Init() with valid config file failed: %v", err)
	}

	config := Get()
	if config.Host != "https://test.example.com" {
		t.Errorf("Expected host 'https://test.example.com', got '%s'", config.Host)
	}
	if config.APIKey != "test-api-key" {
		t.Errorf("Expected api_key 'test-api-key', got '%s'", config.APIKey)
	}
	if config.Site != "test-site" {
		t.Errorf("Expected site 'test-site', got '%s'", config.Site)
	}
	if config.Insecure != false {
		t.Errorf("Expected insecure 'false', got '%v'", config.Insecure)
	}
}

func TestGet(t *testing.T) {
	viper.Reset()
	cfg = nil // Reset the singleton

	viper.Set("host", "https://example.com")
	viper.Set("api_key", "test-key")
	viper.Set("site", "default")
	viper.Set("insecure", true)

	config := Get()

	if config.Host != "https://example.com" {
		t.Errorf("Expected host 'https://example.com', got '%s'", config.Host)
	}
	if config.APIKey != "test-key" {
		t.Errorf("Expected api_key 'test-key', got '%s'", config.APIKey)
	}
	if config.Site != "default" {
		t.Errorf("Expected site 'default', got '%s'", config.Site)
	}
	if config.Insecure != true {
		t.Errorf("Expected insecure 'true', got '%v'", config.Insecure)
	}

	// Test singleton behavior
	config2 := Get()
	if config != config2 {
		t.Error("Get() should return the same instance")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		apiKey    string
		wantError bool
	}{
		{
			name:      "valid config",
			host:      "https://example.com",
			apiKey:    "test-key",
			wantError: false,
		},
		{
			name:      "missing host",
			host:      "",
			apiKey:    "test-key",
			wantError: true,
		},
		{
			name:      "missing api key",
			host:      "https://example.com",
			apiKey:    "",
			wantError: true,
		},
		{
			name:      "both missing",
			host:      "",
			apiKey:    "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			cfg = nil

			viper.Set("host", tt.host)
			viper.Set("api_key", tt.apiKey)

			err := Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	viper.Reset()

	// Test with no config file used
	path := GetConfigPath()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".unifi-cli.yaml")
	if path != expected {
		t.Errorf("Expected config path '%s', got '%s'", expected, path)
	}
}
