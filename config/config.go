package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultNodesPath = "nodes.json"
	defaultStorage   = "storage"
	defaultLogDir    = "logs"
	defaultSSHPort   = 22
	defaultRetention = 7
	defaultTimeout   = 30 * time.Second
	defaultCooldown  = 30 * time.Minute
	defaultPoll      = 5 * time.Second
)

type Config struct {
	NodesPath           string
	StorageDir          string
	LogDir              string
	TelegramBotToken    string
	TelegramChatID      string
	SSHPort             int
	SSHTimeout          time.Duration
	RetentionDays       int
	MaxParallelBackups  int
	AlertCooldown       time.Duration
	MonitorPollInterval time.Duration
	BackupIntervalHours int
	BackupTimezone      string

	WebListen    string
	WebSecret    string
	WebUsersFile string
	WebTLSCert   string
	WebTLSKey    string
	FrontendDir  string
}

func Load(envPath string) (Config, error) {
	if envPath == "" {
		envPath = ".env"
	}

	if err := loadDotEnv(envPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}

	cfg := Config{
		NodesPath:         getEnv("NODES_FILE", defaultNodesPath),
		StorageDir:        getEnv("STORAGE_DIR", defaultStorage),
		LogDir:            getEnv("LOG_DIR", defaultLogDir),
		TelegramBotToken:  strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		TelegramChatID:    strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID")),
		SSHPort:           getEnvInt("SSH_PORT", defaultSSHPort),
		SSHTimeout:        getEnvDuration("SSH_TIMEOUT", defaultTimeout),
		RetentionDays:     getEnvInt("RETENTION_DAYS", defaultRetention),
		MaxParallelBackups:  getEnvInt("MAX_PARALLEL_BACKUPS", 0),
		AlertCooldown:       getEnvDuration("ALERT_COOLDOWN", defaultCooldown),
		MonitorPollInterval: getEnvDuration("MONITOR_POLL_INTERVAL", defaultPoll),
		BackupIntervalHours: getEnvInt("BACKUP_INTERVAL_HOURS", 3),
		BackupTimezone:      getEnv("BACKUP_TIMEZONE", "Asia/Jakarta"),

		WebListen:    getEnv("WEB_LISTEN", "127.0.0.1:8080"),
		WebSecret:    strings.TrimSpace(os.Getenv("WEB_SECRET")),
		WebUsersFile: getEnv("WEB_USERS_FILE", "users.json"),
		WebTLSCert:   strings.TrimSpace(os.Getenv("WEB_TLS_CERT")),
		WebTLSKey:    strings.TrimSpace(os.Getenv("WEB_TLS_KEY")),
		FrontendDir:  getEnv("FRONTEND_DIR", "frontend/dist"),
	}

	if !filepath.IsAbs(cfg.NodesPath) {
		cfg.NodesPath = filepath.Clean(cfg.NodesPath)
	}
	if !filepath.IsAbs(cfg.StorageDir) {
		cfg.StorageDir = filepath.Clean(cfg.StorageDir)
	}
	if !filepath.IsAbs(cfg.LogDir) {
		cfg.LogDir = filepath.Clean(cfg.LogDir)
	}

	if cfg.RetentionDays < 0 {
		cfg.RetentionDays = defaultRetention
	}
	if cfg.AlertCooldown <= 0 {
		cfg.AlertCooldown = defaultCooldown
	}
	if cfg.MonitorPollInterval <= 0 {
		cfg.MonitorPollInterval = defaultPoll
	}

	return cfg, nil
}

func loadDotEnv(path string) error {
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

		key, value, found := strings.Cut(line, "=")
		if !found {
			return fmt.Errorf("invalid env line: %q", line)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)

		if key == "" {
			return fmt.Errorf("invalid env line: %q", line)
		}

		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("set env %s: %w", key, err)
			}
		}
	}

	return scanner.Err()
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
