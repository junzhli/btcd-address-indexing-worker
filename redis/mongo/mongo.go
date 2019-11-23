package mongo

import (
	"encoding/json"
	"time"

	mg "github.com/junzhli/btcd-address-indexing-worker/mongo"

	"github.com/go-redis/redis"
)

// CacheUserHistory returns the data stored in redis
func CacheUserHistory(rs *redis.Client, key string, userdata *mg.UserHistory, ttl time.Duration) error {
	res, err := json.Marshal(*userdata)
	if err != nil {
		return err
	}

	err = rs.Set(key, res, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}

// RestoreUserHistory stores the data to redis
func RestoreUserHistory(rs *redis.Client, key string) (*mg.UserHistory, error) {
	result, err := rs.Get(key).Result()
	if err != nil {
		return nil, err
	}

	var rt mg.UserHistory
	json.Unmarshal([]byte(result), &rt)
	return &rt, nil
}
