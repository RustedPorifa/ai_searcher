package redidb

import (
	"context"
	"log"
	"os"

	redis "github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRedi() {
	log.Println("Starting redis initializing")
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
}
