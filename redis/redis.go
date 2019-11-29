package redis

import (
	"time"

	rs "github.com/go-redis/redis"
)

// Redis deals with Redis data stuff
type Redis interface {
	Get(key string) (string, error)
	Del(key string) error
	Set(key string, value interface{}, expiration time.Duration) error
	Close() error
}

type redis struct {
	client *rs.Client
}

func (r *redis) Get(key string) (string, error) {
	return r.client.Get(key).Result()
}

func (r *redis) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(key, value, expiration).Err()
}

func (r *redis) Del(key string) error {
	return r.client.Del(key).Err()
}

func (r *redis) Close() error {
	return r.client.Close()
}

// New creates an instance of Redis
func New(options *rs.Options) Redis {
	return &redis{
		client: rs.NewClient(options),
	}
}
