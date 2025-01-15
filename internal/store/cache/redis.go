package cache

import "github.com/go-redis/redis/v8"


func NewRedisClient(addrs, pw string, db int) *redis.Client  {
	return redis.NewClient(&redis.Options{
		Addr: addrs,
		Password: pw,
		DB: db,
	})
}