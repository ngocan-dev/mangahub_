package output

import (
	"encoding/json"
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

// PrintJSON pretty prints the payload when verbose mode is enabled.
func PrintJSON(cmd *cobra.Command, payload any) {
	if config.Runtime().Quiet || !config.Runtime().Verbose {
		return
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		cmd.Println(fmt.Sprintf("%v", payload))
		return
	}
	cmd.Println(string(data))
}

// PrintSuccess prints a message unless quiet mode is enabled.
func PrintSuccess(cmd *cobra.Command, message string) {
	if config.Runtime().Quiet {
		return
	}
	cmd.Println(message)
}
