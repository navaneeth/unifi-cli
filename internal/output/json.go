package output

import (
	"encoding/json"
	"fmt"

	"github.com/nkn/unifi-cli/internal/api"
)

func PrintClientsJSON(clients []api.Client) error {
	data, err := json.MarshalIndent(clients, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}
