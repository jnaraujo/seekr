package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jnaraujo/seekr/internal/config"
)

func DefaultStorePath() (string, error) {
	fileName := fmt.Sprintf("store.%s", config.DBFileExt)

	if cfgDir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(cfgDir, config.AppID, fileName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, fmt.Sprintf(".%s", config.AppID), fileName), nil
}

func EnsureStorePath(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0o755)
}

func CheckFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("file %q not found", path)
		}
		return fmt.Errorf("unable to check file %q: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%q is a directory, expected a file", path)
	}
	return nil
}
