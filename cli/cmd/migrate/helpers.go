package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func loadMigrations(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}

	migrations := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		migrations = append(migrations, filepath.Join(dir, entry.Name()))
	}

	sort.Strings(migrations)
	return migrations, nil
}
