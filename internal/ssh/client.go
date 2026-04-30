package sshclient

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"backup-system/config"
	gossh "golang.org/x/crypto/ssh"
)

type Client struct {
	timeout time.Duration
}

type AuthMode string

const (
	AuthModeAuto        AuthMode = "auto"
	AuthModeKeyOnly     AuthMode = "key_only"
	AuthModePasswordOnly AuthMode = "password_only"
)

func New(timeout time.Duration) *Client {
	return &Client{timeout: timeout}
}

func (c *Client) Connect(node config.Node) (*gossh.Client, error) {
	return c.connectWithMode(node, AuthModeAuto)
}

func (c *Client) ConnectKeyOnly(node config.Node) (*gossh.Client, error) {
	return c.connectWithMode(node, AuthModeKeyOnly)
}

func (c *Client) ConnectPasswordOnly(node config.Node) (*gossh.Client, error) {
	return c.connectWithMode(node, AuthModePasswordOnly)
}

func (c *Client) connectWithMode(node config.Node, mode AuthMode) (*gossh.Client, error) {
	authMethods, err := authMethods(node, mode)
	if err != nil {
		return nil, err
	}

	clientConfig := &gossh.ClientConfig{
		User:            node.SSHUser,
		Auth:            authMethods,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         c.timeout,
	}

	address := net.JoinHostPort(node.Host, fmt.Sprintf("%d", node.Port))
	client, err := gossh.Dial("tcp", address, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s: %w", address, err)
	}

	return client, nil
}

func (c *Client) RunCommand(client *gossh.Client, command string) ([]byte, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("create ssh session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return nil, fmt.Errorf("run command %q: %w", command, err)
	}

	return output, nil
}

func (c *Client) RunCommandStream(client *gossh.Client, command string) (io.ReadCloser, <-chan error, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, nil, fmt.Errorf("create ssh session: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("attach stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("attach stderr pipe: %w", err)
	}

	if err := session.Start(command); err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("start remote command %q: %w", command, err)
	}

	waitCh := make(chan error, 1)
	go func() {
		defer close(waitCh)
		defer session.Close()

		errBytes, readErr := io.ReadAll(stderr)
		if readErr != nil {
			waitCh <- fmt.Errorf("read remote stderr: %w", readErr)
			return
		}

		if err := session.Wait(); err != nil {
			stderrText := strings.TrimSpace(string(errBytes))
			if stderrText != "" {
				waitCh <- fmt.Errorf("remote command failed: %w: %s", err, stderrText)
				return
			}
			waitCh <- fmt.Errorf("remote command failed: %w", err)
			return
		}

		if stderrText := strings.TrimSpace(string(errBytes)); stderrText != "" {
			waitCh <- fmt.Errorf("remote command stderr: %s", stderrText)
			return
		}

		waitCh <- nil
	}()

	return readCloser{
		Reader: stdout,
		closeFn: func() error {
			return session.Close()
		},
	}, waitCh, nil
}

type readCloser struct {
	io.Reader
	closeFn func() error
}

func (r readCloser) Close() error {
	if r.closeFn == nil {
		return nil
	}
	return r.closeFn()
}

func IsAuthFailure(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unable to authenticate") ||
		strings.Contains(message, "permission denied") ||
		strings.Contains(message, "no supported methods remain")
}

func authMethods(node config.Node, mode AuthMode) ([]gossh.AuthMethod, error) {
	var methods []gossh.AuthMethod

	keyPath := strings.TrimSpace(node.SSHKey)
	if keyPath != "" && mode != AuthModePasswordOnly {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("read ssh key %s: %w", keyPath, err)
		}

		signer, err := gossh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("parse ssh key %s: %w", keyPath, err)
		}

		methods = append(methods, gossh.PublicKeys(signer))
	}

	password := strings.TrimSpace(node.Password)
	if password != "" && mode != AuthModeKeyOnly {
		methods = append(methods, gossh.Password(password))
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no ssh auth method configured")
	}

	return methods, nil
}
