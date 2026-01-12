package cmd

import (
	"fmt"
	"os"

	"github.com/nkn/unifi-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "unifi",
	Short: "Unifi Network CLI - Manage your Unifi network from the command line",
	Long: `A command-line interface for managing Unifi Network devices using the official Unifi Network API.

This tool allows you to interact with your Unifi controller to manage clients, devices, networks, and more.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return config.Validate()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.unifi-cli.yaml)")
	rootCmd.PersistentFlags().String("host", "", "Unifi controller host (e.g., https://unifi.example.com)")
	rootCmd.PersistentFlags().String("site", "default", "Site ID")
	rootCmd.PersistentFlags().BoolP("insecure", "k", true, "Skip TLS certificate verification")

	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("site", rootCmd.PersistentFlags().Lookup("site"))
	viper.BindPFlag("insecure", rootCmd.PersistentFlags().Lookup("insecure"))
}

func initConfig() {
	if err := config.Init(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}
}
