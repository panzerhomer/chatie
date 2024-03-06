package redis

import (
	"log"

	"github.com/redis/go-redis/v9"
)

func CreateRedisClient() *redis.Client {
	opt, err := redis.ParseURL("redis://localhost:6364/0")
	if err != nil {
		log.Println(err)
	}

	return redis.NewClient(opt)
}
