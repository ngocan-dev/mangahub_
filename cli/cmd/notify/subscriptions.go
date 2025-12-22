package notify

import (
	"errors"
	"sort"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
)

func normalizeSubscriptionID(id string) (string, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return "", errors.New("novelID cannot be empty")
	}
	return trimmed, nil
}

func addSubscription(cfg *config.Manager, novelID string) (bool, error) {
	if cfg == nil {
		return false, errors.New("configuration not loaded")
	}

	for _, existing := range cfg.Data.Notifications.Subscriptions {
		if strings.EqualFold(existing, novelID) {
			return false, nil
		}
	}

	cfg.Data.Notifications.Subscriptions = append(cfg.Data.Notifications.Subscriptions, novelID)
	sort.Strings(cfg.Data.Notifications.Subscriptions)
	return true, cfg.Save()
}

func removeSubscription(cfg *config.Manager, novelID string) (bool, error) {
	if cfg == nil {
		return false, errors.New("configuration not loaded")
	}

	updated := false
	remaining := cfg.Data.Notifications.Subscriptions[:0]
	for _, existing := range cfg.Data.Notifications.Subscriptions {
		if strings.EqualFold(existing, novelID) {
			updated = true
			continue
		}
		remaining = append(remaining, existing)
	}

	cfg.Data.Notifications.Subscriptions = remaining
	if !updated {
		return false, nil
	}
	return true, cfg.Save()
}

func listSubscriptions(cfg *config.Manager) []string {
	if cfg == nil {
		return nil
	}
	return append([]string{}, cfg.Data.Notifications.Subscriptions...)
}
