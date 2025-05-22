package storage

import (
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
