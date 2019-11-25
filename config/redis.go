package config

import "os"

// Names
const (
	RedisHost     string = "REDIS_HOST"
	RedisPassword string = "REDIS_PASSWORD"
)

// Default values
const (
	DefaultRedisHost     string = "127.0.0.1:6379"
	DefaultRedisPassword string = ""
)

// RedisConfig prepared for runtime environment
type RedisConfig struct {
	Host     string
	Password string
}

// LoadRedisConfig returns RedisConfig
func LoadRedisConfig() (*RedisConfig, error) {
	host := os.Getenv(RedisHost)
	if host == "" {
		EmptyOnLoad(RedisHost, true, DefaultRedisHost)
		host = DefaultRedisHost
	}

	pass := os.Getenv(RedisPassword)
	if pass == "" {
		EmptyOnLoad(RedisPassword, true, DefaultRedisPassword)
		pass = DefaultRedisPassword
	}

	return &RedisConfig{
		Host:     host,
		Password: pass,
	}, nil
}
