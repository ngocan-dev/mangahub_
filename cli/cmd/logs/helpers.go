package logs

import "github.com/ngocan-dev/mangahub_/cli/internal/config"

func quiet() bool {
	return config.Runtime().Quiet
}
