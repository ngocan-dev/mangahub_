package export

import "strings"

func formatUpper(format string) string {
	if format == "" {
		return ""
	}
	return strings.ToUpper(format)
}
