package output

import (
	"os"

	"github.com/nkn/unifi-cli/internal/api"
	"github.com/olekukonko/tablewriter"
)

func PrintClientsTable(clients []api.Client) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Append([]string{"MAC", "Name", "IP", "Type", "SSID", "Signal", "Uptime", "RX/TX"})

	for _, client := range clients {
		rxTx := api.FormatBytes(client.RxBytes) + " / " + api.FormatBytes(client.TxBytes)

		row := []string{
			client.MAC,
			client.GetDisplayName(),
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
