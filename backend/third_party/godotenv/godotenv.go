package godotenv

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Load reads the provided .env files (or ".env" when none are specified)
// and sets environment variables for any keys found. Existing values are
// left untouched to mimic the upstream library's behavior.
func Load(filenames ...string) error {
	if len(filenames) == 0 {
		filenames = []string{".env"}
	}

	for _, name := range filenames {
		if err := loadFile(name); err != nil {
			return err
		}
	}
	return nil
}

func loadFile(name string) error {
	path := name
	if !filepath.IsAbs(name) {
		if cwd, err := os.Getwd(); err == nil {
			path = filepath.Join(cwd, name)
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, strings.Trim(val, `"'`)); err != nil {
				return fmt.Errorf("set env %s: %w", key, err)
			}
		}
	}

	return scanner.Err()
}
