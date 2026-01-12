package cmd

import (
	"fmt"
	"strings"

	"github.com/nkn/unifi-cli/internal/api"
	"github.com/nkn/unifi-cli/internal/config"
	"github.com/nkn/unifi-cli/internal/filter"
	"github.com/nkn/unifi-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	outputFormat   string
	filterWired    bool
	filterWireless bool
	filterBlocked  bool
	filterAP       string
	filterSQL      string
)

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
	clientsListCmd.Flags().BoolVar(&filterWired, "wired", false, "Show only wired clients")
	clientsListCmd.Flags().BoolVar(&filterWireless, "wireless", false, "Show only wireless clients")
	clientsListCmd.Flags().BoolVar(&filterBlocked, "blocked", false, "Show only blocked clients")
	clientsListCmd.Flags().StringVar(&filterAP, "ap", "", "Filter by Access Point MAC address")
	clientsListCmd.Flags().StringVar(&filterSQL, "filter", "", "SQL WHERE clause (e.g., 'signal >= -65 AND essid = \"HomeWiFi\"')")
}

func runClientsList(cmd *cobra.Command, args []string) error {
	cfg := config.Get()

	apiClient := api.NewAPIClient(cfg.Host, cfg.APIKey, cfg.Site, cfg.Insecure)

	clients, err := apiClient.ListClients()
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}

	// Build WHERE clause from flags
	whereClause, err := buildWhereClause()
	if err != nil {
		return err
	}

	// Apply filter if needed
	filteredClients := clients
	if whereClause != "" {
		filterEngine, err := filter.NewFilter(whereClause)
		if err != nil {
			return fmt.Errorf("failed to create filter: %w", err)
		}
		defer filterEngine.Close()

		filteredClients, err = filterEngine.Apply(clients)
		if err != nil {
			return fmt.Errorf("failed to apply filter: %w", err)
		}
	}

	if len(filteredClients) == 0 {
		fmt.Println("No clients match the specified filters")
		return nil
	}

	switch outputFormat {
	case "json":
		return output.PrintClientsJSON(filteredClients)
	case "table":
		output.PrintClientsTable(filteredClients)
		return nil
	default:
		return fmt.Errorf("invalid output format: %s (valid options: table, json)", outputFormat)
	}
}

func buildWhereClause() (string, error) {
	var conditions []string

	// Validate mutually exclusive flags
	if filterWired && filterWireless {
		return "", fmt.Errorf("--wired and --wireless are mutually exclusive")
	}

	// Build conditions from simple flags
	if filterWired {
		conditions = append(conditions, "is_wired = 1")
	}
	if filterWireless {
		conditions = append(conditions, "is_wired = 0")
	}
	if filterBlocked {
		conditions = append(conditions, "blocked = 1")
	}
	if filterAP != "" {
		conditions = append(conditions, fmt.Sprintf("ap_mac = '%s'", filterAP))
	}

	// Add custom SQL filter
	if filterSQL != "" {
		conditions = append(conditions, fmt.Sprintf("(%s)", filterSQL))
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return strings.Join(conditions, " AND "), nil
}
