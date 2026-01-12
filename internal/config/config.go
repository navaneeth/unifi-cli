package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Host     string
	APIKey   string
	Site     string
	Insecure bool
}

var cfg *Config

func Init(cfgFile string) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".unifi-cli")
	}

	viper.SetEnvPrefix("UNIFI")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("site", "default")
	viper.SetDefault("insecure", true)

	// Read config file (if it exists)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return nil
}

func Get() *Config {
	if cfg == nil {
		cfg = &Config{
			Host:     viper.GetString("host"),
			APIKey:   viper.GetString("api_key"),
			Site:     viper.GetString("site"),
			Insecure: viper.GetBool("insecure"),
		}
	}
	return cfg
}

func Validate() error {
	cfg := Get()

	if cfg.Host == "" {
		return fmt.Errorf("host is required (set via --host, UNIFI_HOST, or config file)")
	}

	if cfg.APIKey == "" {
		return fmt.Errorf("API key is required (set via UNIFI_API_KEY or config file)")
	}

	return nil
}

func GetConfigPath() string {
	if viper.ConfigFileUsed() != "" {
		return viper.ConfigFileUsed()
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".unifi-cli.yaml")
}
