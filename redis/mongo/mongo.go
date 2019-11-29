package mongo

import (
	"encoding/json"
	mg "github.com/junzhli/btcd-address-indexing-worker/mongo"
	"github.com/junzhli/btcd-address-indexing-worker/redis"
	"time"
)

// CacheUserHistory stores the data to redis
func CacheUserHistory(rs redis.Redis, key string, userdata *mg.UserHistory, ttl time.Duration) error {
	res, err := json.Marshal(*userdata)
	if err != nil {
		return err
	}

	err = rs.Set(key, res, ttl)
	if err != nil {
		return err
	}
	return nil
}

// RestoreUserHistory returns the data stored in redis
func RestoreUserHistory(rs redis.Redis, key string) (*mg.UserHistory, error) {
	result, err := rs.Get(key)
	if err != nil {
		return nil, err
	}

	var rt mg.UserHistory
	json.Unmarshal([]byte(result), &rt)
	return &rt, nil
}
