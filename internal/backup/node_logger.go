package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type NodeLogger struct {
	baseDir string
	mu      sync.Mutex
}

func NewNodeLogger(baseDir string) *NodeLogger {
	return &NodeLogger{baseDir: baseDir}
}

func (l *NodeLogger) Log(nodeName, status, message string) error {
	if err := os.MkdirAll(l.baseDir, 0o755); err != nil {
		return fmt.Errorf("create log dir %s: %w", l.baseDir, err)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	path := filepath.Join(l.baseDir, sanitizeLogName(nodeName)+".log")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file %s: %w", path, err)
	}
	defer file.Close()

	line := fmt.Sprintf("[%s] %s %s\n", time.Now().UTC().Format("2006-01-02 15:04:05"), status, message)
	if _, err := file.WriteString(line); err != nil {
		return fmt.Errorf("write log file %s: %w", path, err)
	}

	return nil
}

func sanitizeLogName(value string) string {
	name := filepath.Clean(value)
	name = filepath.Base(name)
	if name == "." || name == string(filepath.Separator) || name == "" {
		return "unknown"
	}
	return name
}
