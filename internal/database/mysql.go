package database

import (
	"fmt"
	"io"
	"strings"
	"time"

	"backup-system/config"
	sshclient "backup-system/internal/ssh"
)

type MySQLBackup struct {
	ssh *sshclient.Client
}

func NewMySQLBackup(ssh *sshclient.Client) *MySQLBackup {
	return &MySQLBackup{ssh: ssh}
}

func (b *MySQLBackup) BuildDumpCommand(node config.Node) string {
	return fmt.Sprintf(
		"mysqldump --single-transaction --quick --routines --triggers --user=%s --password=%s %s | gzip -c",
		shellEscape(node.DBUser),
		shellEscape(node.DBPass),
		shellEscape(node.DBName),
	)
}

func (b *MySQLBackup) Stream(node config.Node) (io.ReadCloser, <-chan error, error) {
	client, err := b.ssh.Connect(node)
	if err != nil {
		return nil, nil, err
	}

	stream, waitCh, err := b.ssh.RunCommandStream(client, b.BuildDumpCommand(node))
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	doneCh := make(chan error, 1)
	go func() {
		defer close(doneCh)
		defer client.Close()
		doneCh <- <-waitCh
	}()

	return stream, doneCh, nil
}

func BackupFilename(name string, now time.Time) string {
	return fmt.Sprintf("%s-%s.sql.gz", sanitizeFileName(name), now.Format("2006-01-02"))
}

func shellEscape(value string) string {
	if value == "" {
		return "''"
	}

	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func sanitizeFileName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "-")
	return strings.ToLower(value)
}
