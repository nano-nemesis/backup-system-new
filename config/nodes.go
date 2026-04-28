package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	NodeTypeMikrotik = "mikrotik"
	NodeTypeDatabase = "database"
)

type Node struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Host       string `json:"host"`
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`

	SSHUser string `json:"ssh_user"`
	SSHKey  string `json:"ssh_key"`

	DBName string `json:"db_name"`
	DBUser string `json:"db_user"`
	DBPass string `json:"db_pass"`
}

func LoadNodes(path string) ([]Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read nodes file: %w", err)
	}

	var nodes []Node
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, fmt.Errorf("parse nodes file: %w", err)
	}

	for i := range nodes {
		nodes[i].normalize()
		if err := nodes[i].Validate(); err != nil {
			return nil, fmt.Errorf("node %q: %w", nodes[i].Name, err)
		}
	}

	return nodes, nil
}

func (n *Node) normalize() {
	n.Name = strings.TrimSpace(n.Name)
	n.Type = strings.ToLower(strings.TrimSpace(n.Type))
	if n.Type == "" {
		n.Type = NodeTypeMikrotik
	}

	if strings.TrimSpace(n.Host) == "" {
		n.Host = strings.TrimSpace(n.IP)
	}

	if strings.TrimSpace(n.SSHUser) == "" {
		n.SSHUser = strings.TrimSpace(n.User)
	}

	if strings.TrimSpace(n.SSHKey) == "" {
		n.SSHKey = strings.TrimSpace(n.PrivateKey)
	}
}

func (n Node) Validate() error {
	if n.Name == "" {
		return fmt.Errorf("name is required")
	}

	if n.Host == "" {
		return fmt.Errorf("host is required")
	}

	switch n.Type {
	case NodeTypeMikrotik:
		if n.User == "" && n.SSHUser == "" {
			return fmt.Errorf("user is required for mikrotik backup")
		}
		if strings.TrimSpace(n.Password) == "" && strings.TrimSpace(n.SSHKey) == "" && strings.TrimSpace(n.PrivateKey) == "" {
			return fmt.Errorf("password or ssh key is required for mikrotik backup")
		}
	case NodeTypeDatabase:
		if n.SSHUser == "" {
			return fmt.Errorf("ssh_user is required for database backup")
		}
		if strings.TrimSpace(n.SSHKey) == "" && strings.TrimSpace(n.Password) == "" {
			return fmt.Errorf("ssh_key or password is required for database backup")
		}
		if n.DBName == "" || n.DBUser == "" {
			return fmt.Errorf("db_name and db_user are required for database backup")
		}
	default:
		return fmt.Errorf("unsupported type %q", n.Type)
	}

	return nil
}
