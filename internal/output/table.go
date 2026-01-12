package output

import (
	"fmt"
	"os"

	"github.com/nkn/unifi-cli/internal/api"
	"github.com/olekukonko/tablewriter"
)

func PrintClientsTable(clients []api.Client) {
	table := tablewriter.NewWriter(os.Stdout)

	// Add header row
	table.Append([]string{"Name", "IP", "Type", "SSID", "Signal", "Uptime", "RX/TX"})

	for _, client := range clients {
		rxTx := api.FormatBytes(client.RxBytes) + " / " + api.FormatBytes(client.TxBytes)

		// Combine name and MAC address - MAC shown in parentheses to save space
		nameWithMAC := fmt.Sprintf("%s (%s)", client.GetDisplayName(), client.MAC)

		row := []string{
			nameWithMAC,
			client.IP,
			client.GetConnectionType(),
			client.GetSSID(),
			client.GetSignal(),
			client.GetUptime(),
			rxTx,
		}

		table.Append(row)
	}

	table.Render()
}
