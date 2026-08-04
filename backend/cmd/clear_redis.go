package main

import (
	"context"
	"fmt"

	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadConfig()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
	})

	err := rdb.FlushAll(context.Background()).Err()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Redis cleared")
	}
}