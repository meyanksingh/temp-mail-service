package redis

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

var Error error

func InitRedis() {

	dsn := os.Getenv("REDIS_URL")
	if dsn == "" {
		log.Fatal("Error Loading the ENV File")
	}

	opt, err := redis.ParseURL(dsn)
	if err != nil {
		log.Fatalf("Error Parsing the URL: %s", err)
	}
	RedisClient = redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully!")
}
