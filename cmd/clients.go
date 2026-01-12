package cmd

import (
	"fmt"

	"github.com/nkn/unifi-cli/internal/api"
	"github.com/nkn/unifi-cli/internal/config"
	"github.com/nkn/unifi-cli/internal/output"
	"github.com/spf13/cobra"
)

var outputFormat string

var clientsCmd = &cobra.Command{
	Use:   "clients",
	Short: "Manage Unifi clients",
	Long:  `View and manage connected clients on your Unifi network.`,
}

var clientsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List connected clients",
	Long:  `List all currently connected clients on the Unifi network.`,
	RunE:  runClientsList,
}

func init() {
	rootCmd.AddCommand(clientsCmd)
	clientsCmd.AddCommand(clientsListCmd)

	clientsListCmd.Flags().StringVarP(&outputFormat, "format", "f", "table", "Output format (table or json)")
}

func runClientsList(cmd *cobra.Command, args []string) error {
	cfg := config.Get()

	apiClient := api.NewAPIClient(cfg.Host, cfg.APIKey, cfg.Site, cfg.Insecure)

	clients, err := apiClient.ListClients()
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}

	if len(clients) == 0 {
		fmt.Println("No connected clients found")
		return nil
	}

	switch outputFormat {
	case "json":
		return output.PrintClientsJSON(clients)
	case "table":
		output.PrintClientsTable(clients)
		return nil
	default:
		return fmt.Errorf("invalid output format: %s (valid options: table, json)", outputFormat)
	}
}
