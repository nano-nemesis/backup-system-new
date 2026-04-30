package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Storage struct {
	baseDir string
}

func NewStorage(baseDir string) *Storage {
	return &Storage{baseDir: baseDir}
}

func (s *Storage) Save(category, filename string, reader io.Reader) (string, int64, error) {
	targetDir := filepath.Join(s.baseDir, category)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", 0, fmt.Errorf("create storage dir %s: %w", targetDir, err)
	}

	targetPath := filepath.Join(targetDir, filename)
	file, err := os.Create(targetPath)
	if err != nil {
		return "", 0, fmt.Errorf("create backup file %s: %w", targetPath, err)
	}

	written, copyErr := io.Copy(file, reader)
	closeErr := file.Close()

	if copyErr != nil {
		_ = os.Remove(targetPath)
		return "", 0, fmt.Errorf("write backup file %s: %w", targetPath, copyErr)
	}
	if closeErr != nil {
		return "", 0, fmt.Errorf("close backup file %s: %w", targetPath, closeErr)
	}

	return targetPath, written, nil
}

func (s *Storage) PurgeOlderThan(category string, retentionDays int, now time.Time) error {
	if retentionDays <= 0 {
		return nil
	}

	targetDir := filepath.Join(s.baseDir, category)
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read storage dir %s: %w", targetDir, err)
	}

	cutoff := now.AddDate(0, 0, -retentionDays)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("stat file %s: %w", entry.Name(), err)
		}

		if info.ModTime().Before(cutoff) {
			if err := os.Remove(filepath.Join(targetDir, entry.Name())); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove old backup %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}
