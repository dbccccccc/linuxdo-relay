package storage

import "github.com/go-redis/redis/v8"

type Redis struct {
	*redis.Client
}

func NewRedis(addr, password string) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &Redis{Client: client}
}
