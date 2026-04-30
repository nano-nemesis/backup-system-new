package web

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"backup-system/config"
	"backup-system/internal/utils"
)

type NodeStatus struct {
	Name        string
	Type        string
	Host        string
	LastStatus  string
	LastTime    time.Time
	LastMsg     string
	BackupFiles []BackupFile
}

type BackupFile struct {
	Name    string
	RelPath string
	Size    int64
	SizeStr string
	ModTime time.Time
}

func ReadStatuses(cfg config.Config) ([]NodeStatus, error) {
	nodes, err := config.LoadNodes(cfg.NodesPath)
	if err != nil {
		return nil, err
	}

	out := make([]NodeStatus, 0, len(nodes))
	for _, n := range nodes {
		ns := NodeStatus{
			Name: n.Name,
			Type: n.Type,
			Host: n.Host,
		}
		ns.LastStatus, ns.LastTime, ns.LastMsg = readLastStatus(cfg.LogDir, n.Name)
		ns.BackupFiles = listBackupFiles(cfg.StorageDir, n.Type, n.Name)
		out = append(out, ns)
	}
	return out, nil
}

func readLastStatus(logDir, nodeName string) (string, time.Time, string) {
	path := filepath.Join(logDir, safeLogName(nodeName)+".log")
	f, err := os.Open(path)
	if err != nil {
		return "UNKNOWN", time.Time{}, ""
	}
	defer f.Close()

	var lastStatus, lastMsg string
	var lastTime time.Time

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		t, status, msg := parseLine(line)
		if status != "" {
			lastTime = t
			lastStatus = status
			lastMsg = msg
		}
	}

	if lastStatus == "" {
		return "UNKNOWN", time.Time{}, ""
	}
	return lastStatus, lastTime, lastMsg
}

// parseLine parses: [2006-01-02 15:04:05] STATUS message
func parseLine(line string) (time.Time, string, string) {
	if len(line) < 22 || line[0] != '[' {
		return time.Time{}, "", ""
	}
	end := strings.Index(line, "]")
	if end < 0 {
		return time.Time{}, "", ""
	}
	t, err := time.Parse("2006-01-02 15:04:05", line[1:end])
	if err != nil {
		return time.Time{}, "", ""
	}
	rest := strings.TrimSpace(line[end+1:])
	switch {
	case strings.HasPrefix(rest, "SUCCESS"):
		return t, "SUCCESS", rest
	case strings.HasPrefix(rest, "ERROR"):
		return t, "ERROR", rest
	}
	return time.Time{}, "", ""
}

func listBackupFiles(storageDir, nodeType, nodeName string) []BackupFile {
	dir := filepath.Join(storageDir, nodeType)
	prefix := safeLogName(nodeName) + "-"

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var files []BackupFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, BackupFile{
			Name:    e.Name(),
			RelPath: filepath.Join(nodeType, e.Name()),
			Size:    info.Size(),
			SizeStr: utils.FormatBytes(info.Size()),
			ModTime: info.ModTime(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})
	return files
}

func safeLogName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "-")
	return strings.ToLower(name)
}
