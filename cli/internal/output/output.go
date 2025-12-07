package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

const (
	// FormatText represents the standard human-readable output.
	FormatText = "text"
	// FormatJSON renders structured JSON output.
	FormatJSON = "json"
	// OutputFlag defines the shared flag name for selecting output format.
	OutputFlag = "output"
)

// AddFlag registers a shared --output flag for read-only commands.
func AddFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(OutputFlag, "o", FormatText, "Output format (text|json)")
}

// GetFormat resolves the requested output format, validating supported options.
func GetFormat(cmd *cobra.Command) (string, error) {
	if cmd == nil {
		return FormatText, nil
	}
	flag := cmd.Flags().Lookup(OutputFlag)
	if flag == nil {
		return FormatText, nil
	}
	value := strings.ToLower(strings.TrimSpace(flag.Value.String()))
	if value == "" {
		value = FormatText
	}
	switch value {
	case FormatText, FormatJSON:
		return value, nil
	default:
		return "", fmt.Errorf("✗ Invalid output format: %s\nSupported formats: text, json", value)
	}
}

// PrintJSON pretty prints payloads when JSON output is requested or verbose mode is enabled.
func PrintJSON(cmd *cobra.Command, payload any) {
	format, _ := GetFormat(cmd)
	if config.Runtime().Quiet || (format != FormatJSON && !config.Runtime().Verbose) {
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
