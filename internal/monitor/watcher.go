package monitor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Watcher struct {
	dir          string
	pollInterval time.Duration
	logger       Logger
}

type Logger interface {
	Printf(format string, v ...any)
}

func NewWatcher(dir string, pollInterval time.Duration, logger Logger) *Watcher {
	return &Watcher{
		dir:          dir,
		pollInterval: pollInterval,
		logger:       logger,
	}
}

func (w *Watcher) Watch(ctx context.Context, handler func(Event)) error {
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return fmt.Errorf("create monitor dir %s: %w", w.dir, err)
	}

	files, err := filepath.Glob(filepath.Join(w.dir, "*.log"))
	if err != nil {
		return fmt.Errorf("list monitor log files: %w", err)
	}

	var wg sync.WaitGroup
	seen := make(map[string]struct{})

	startWatch := func(path string) {
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			w.followFile(ctx, path, handler)
		}()
	}

	for _, path := range files {
		startWatch(path)
	}

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return nil
		case <-ticker.C:
			paths, err := filepath.Glob(filepath.Join(w.dir, "*.log"))
			if err != nil {
				w.logger.Printf("monitor glob failed: %v", err)
				continue
			}
			for _, path := range paths {
				startWatch(path)
			}
		}
	}
}

func (w *Watcher) followFile(ctx context.Context, path string, handler func(Event)) {
	file, err := os.Open(path)
	if err != nil {
		w.logger.Printf("monitor open file failed %s: %v", path, err)
		return
	}
	defer file.Close()

	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		w.logger.Printf("monitor seek file failed %s: %v", path, err)
		return
	}

	reader := bufio.NewReader(file)
	nodeName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(w.pollInterval)
				continue
			}
			w.logger.Printf("monitor read file failed %s: %v", path, err)
			time.Sleep(w.pollInterval)
			continue
		}

		event, ok := ParseLine(nodeName, line)
		if ok {
			handler(event)
		}
	}
}
