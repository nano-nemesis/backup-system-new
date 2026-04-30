package mikrotik

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"backup-system/config"
	sshclient "backup-system/internal/ssh"
	gossh "golang.org/x/crypto/ssh"
)

type Backup struct {
	ssh *sshclient.Client
}

type FetchResult struct {
	Reader               io.Reader
	UsedPasswordFallback bool
}

func NewBackup(ssh *sshclient.Client) *Backup {
	return &Backup{ssh: ssh}
}

func (b *Backup) Fetch(node config.Node) (FetchResult, error) {
	client, usedPasswordFallback, err := b.connect(node)
	if err != nil {
		return FetchResult{}, err
	}
	defer client.Close()

	command := "/export terse"
	output, err := b.ssh.RunCommand(client, command)
	if err != nil {
		return FetchResult{}, fmt.Errorf("mikrotik export failed: %w", err)
	}

	return FetchResult{
		Reader:               bytes.NewReader(output),
		UsedPasswordFallback: usedPasswordFallback,
	}, nil
}

func BackupFilename(name string, now time.Time) string {
	return fmt.Sprintf("%s-%s.rsc", sanitizeFileName(name), now.Format("2006-01-02"))
}

func sanitizeFileName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "-")
	return strings.ToLower(value)
}

func (b *Backup) connect(node config.Node) (*gossh.Client, bool, error) {
	if strings.TrimSpace(node.SSHKey) != "" {
		client, err := b.ssh.ConnectKeyOnly(node)
		if err == nil {
			return client, false, nil
		}

		if sshclient.IsAuthFailure(err) && strings.TrimSpace(node.Password) != "" {
			client, passwordErr := b.ssh.ConnectPasswordOnly(node)
			if passwordErr == nil {
				return client, true, nil
			}
			return nil, false, fmt.Errorf("ssh key auth failed (%v), password fallback failed: %w", err, passwordErr)
		}

		return nil, false, err
	}

	client, err := b.ssh.ConnectPasswordOnly(node)
	if err != nil {
		return nil, false, err
	}

	return client, false, nil
}
