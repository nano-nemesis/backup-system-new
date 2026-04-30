package monitor

import "strings"

const (
	StatusError   = "ERROR"
	StatusSuccess = "SUCCESS"
)

type Event struct {
	NodeName string
	Status   string
	Line     string
	Category string
	Action   string
}

func ParseLine(nodeName, line string) (Event, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return Event{}, false
	}

	switch {
	case strings.Contains(trimmed, StatusError):
		category, action := classify(trimmed)
		return Event{
			NodeName: nodeName,
			Status:   StatusError,
			Line:     trimmed,
			Category: category,
			Action:   action,
		}, true
	case strings.Contains(trimmed, StatusSuccess):
		return Event{
			NodeName: nodeName,
			Status:   StatusSuccess,
			Line:     trimmed,
			Category: "RECOVERY",
			Action:   "backup kembali normal",
		}, true
	default:
		return Event{}, false
	}
}

func classify(line string) (string, string) {
	value := strings.ToLower(line)

	switch {
	case strings.Contains(value, "timeout"), strings.Contains(value, "i/o timeout"):
		return "NETWORK", "cek firewall, latency, atau koneksi SSH"
	case strings.Contains(value, "permission denied"), strings.Contains(value, "unable to authenticate"):
		return "AUTH", "cek credential, SSH key, atau izin akses"
	case strings.Contains(value, "no such file"), strings.Contains(value, "not found"):
		return "DEPENDENCY", "cek path ssh key, mysqldump, atau gzip di remote host"
	case strings.Contains(value, "mysqldump"), strings.Contains(value, "access denied"), strings.Contains(value, "unknown database"):
		return "DATABASE", "cek user backup, password, dan akses database"
	default:
		return "UNKNOWN", "cek detail error pada log node"
	}
}
