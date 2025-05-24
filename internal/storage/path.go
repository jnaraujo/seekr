package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

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

type PathKind int

const (
	InvalidPathKind PathKind = iota
	DirectoryPathKind
	FilePathKind
)

func CheckPath(path string) (PathKind, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return InvalidPathKind, fmt.Errorf("path %q not found", path)
		}
		return InvalidPathKind, fmt.Errorf("unable to check path %q: %w", path, err)
	}

	if fi.IsDir() {
		return DirectoryPathKind, nil
	}
	return FilePathKind, nil
}

func FilePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func IsFileValid(content []byte) bool {
	return utf8.Valid(content)
}
