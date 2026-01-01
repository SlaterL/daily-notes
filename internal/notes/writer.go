package notes

import (
	"os"
	"path/filepath"

	"github.com/SlaterL/daily-notes/internal/config"
)

func DailyNotePath(cfg *config.Config, date string) (string, error) {
	dir := filepath.Join(cfg.VaultPath, cfg.DailyNotesSubdir)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	return filepath.Join(dir, date+".md"), nil
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func Write(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func Append(path string, content []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(content); err != nil {
		return err
	}
	return nil
}
