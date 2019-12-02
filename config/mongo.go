package config

import (
	"os"
)

// Names
const (
	MongoHost     string = "MONGO_HOST"
	MongoUser     string = "MONGO_USER"
	MongoPassword string = "MONGO_PASSWORD"
)

// Default values
const (
	DefaultMongoHost     string = "127.0.0.1:27017"
	DefaultMongoUser     string = ""
	DefaultMongoPassword string = ""
)

// MongoConfig prepared for runtime environment
type MongoConfig struct {
	Host     string
	Username string
	Password string
}

// GetConnectionString returns url represented as connection string
func (m *MongoConfig) GetConnectionString() string {
	if m.Username == "" || m.Password == "" {
		return m.Host
	}
	return m.Username + ":" + m.Password + "@" + m.Host
}

// LoadMongoConfig returns MongoConfig
func LoadMongoConfig() (*MongoConfig, error) {
	host := os.Getenv(MongoHost)
	if host == "" {
		EmptyOnLoad(MongoHost, true, DefaultMongoHost)
		host = DefaultMongoHost
	}

	user := os.Getenv(MongoUser)
	if user == "" {
		EmptyOnLoad(MongoUser, true, DefaultMongoUser)
		user = DefaultMongoUser
	}

	pass := os.Getenv(MongoPassword)
	if pass == "" {
		EmptyOnLoad(MongoPassword, true, DefaultMongoPassword)
		pass = DefaultMongoPassword
	}

	return &MongoConfig{
		Host:     host,
		Username: user,
		Password: pass,
	}, nil
}
