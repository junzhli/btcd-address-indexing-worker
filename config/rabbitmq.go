package config

import "os"

// Names
const (
	RabbitMQHost     string = "RABBITMQ_HOST"
	RabbitMQUser     string = "RABBITMQ_USER"
	RabbitMQPassword string = "RABBITMQ_PASSWORD"
)

// Default values
const (
	DefaultRabbitMQHost     string = "127.0.0.1:5672"
	DefaultRabbitMQUser     string = "guest"
	DefaultRabbitMQPassword string = "guest"
)

// RabbitMQConfig prepared for runtime environment
type RabbitMQConfig struct {
	Host     string
	Username string
	Password string
}

// GetConnectionString returns url represented as connection string
func (m *RabbitMQConfig) GetConnectionString() string {
	if m.Username == "" || m.Password == "" {
		return m.Host
	}
	return m.Username + ":" + m.Password + "@" + m.Host
}

// LoadRabbitMQConfig returns RabbitMQConfig
func LoadRabbitMQConfig() (*RabbitMQConfig, error) {
	host := os.Getenv(RabbitMQHost)
	if host == "" {
		EmptyOnLoad(RabbitMQHost, true, DefaultRabbitMQHost)
		host = DefaultRabbitMQHost
	}

	user := os.Getenv(RabbitMQUser)
	if user == "" {
		EmptyOnLoad(RabbitMQUser, true, DefaultRabbitMQUser)
		user = DefaultRabbitMQUser
	}

	pass := os.Getenv(RabbitMQPassword)
	if pass == "" {
		EmptyOnLoad(RabbitMQPassword, true, DefaultRabbitMQPassword)
		pass = DefaultRabbitMQPassword
	}

	return &RabbitMQConfig{
		Host:     host,
		Username: user,
		Password: pass,
	}, nil
}
