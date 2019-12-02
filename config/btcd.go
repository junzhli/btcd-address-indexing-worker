package config

import (
	"os"
	"strconv"
)

// Names
const (
	BtcdJSONRPCHost     string = "BTCD_JSONRPC_HOST"
	BtcdJSONRPCUser     string = "BTCD_JSONRPC_USER"
	BtcdJSONRPCPassword string = "BTCD_JSONRPC_PASSWORD"
	BtcdJSONRPCTimeout  string = "BTCD_JSONRPC_TIMEOUT"
)

// Default values
const (
	DefaultBtcdJSONRPCHost     string = "127.0.0.1:8334"
	DefaultBtcdJSONRPCUser     string = ""
	DefaultBtcdJSONRPCPassword string = ""
	DefaultBtcdJSONRPCTimeout  int64  = 600
)

// BtcdConfig prepared for runtime environment
type BtcdConfig struct {
	Host     string
	Username string
	Password string
	Timeout  int64
}

// LoadBtcdConfig returns BtcdConfig
func LoadBtcdConfig() (*BtcdConfig, error) {
	host := os.Getenv(BtcdJSONRPCHost)
	if host == "" {
		EmptyOnLoad(BtcdJSONRPCHost, true, DefaultBtcdJSONRPCHost)
		host = DefaultBtcdJSONRPCHost
	}

	user := os.Getenv(BtcdJSONRPCUser)
	if user == "" {
		EmptyOnLoad(BtcdJSONRPCUser, true, DefaultBtcdJSONRPCUser)
		user = DefaultBtcdJSONRPCUser
	}

	pass := os.Getenv(BtcdJSONRPCPassword)
	if pass == "" {
		EmptyOnLoad(BtcdJSONRPCPassword, true, DefaultBtcdJSONRPCPassword)
		pass = DefaultBtcdJSONRPCPassword
	}

	timeout, err := strconv.ParseInt(os.Getenv(BtcdJSONRPCTimeout), 10, 64)
	if err != nil {
		EmptyOnLoad(BtcdJSONRPCTimeout, true, strconv.FormatInt(DefaultBtcdJSONRPCTimeout, 10))
		timeout = DefaultBtcdJSONRPCTimeout
	}

	return &BtcdConfig{
		Host:     host,
		Username: user,
		Password: pass,
		Timeout:  timeout,
	}, nil
}
