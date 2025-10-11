package redidb

import (
	"ai_tg_search/struct_types/newstypes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	redis "github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func InitRedi() {
	log.Println("Starting redis initializing")
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	if err := Rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
}

func CacheNews(topic string, news *newstypes.NewsResponse, ttl time.Duration) error {
	ctx := context.Background()
	data, _ := json.Marshal(news)
	return Rdb.Set(ctx, "news:"+topic, data, ttl).Err()
}

func AddToHistory(userID int64, topic string) {
	ctx := context.Background()
	key := fmt.Sprintf("user:%d:news_history", userID)
	Rdb.LRem(ctx, key, 1, topic) // убрать дубль
	Rdb.LPush(ctx, key, topic)   // добавить в начало
	Rdb.LTrim(ctx, key, 0, 9)    // оставить 10

	Rdb.Expire(ctx, key, time.Hour*24)
}

func GetUserTopics(userID int64) ([]string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("user:%d:news_history", userID)

	topics, err := Rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user topics: %w", err)
	}

	return topics, nil
}

func GetNewsByTopic(topic string) (*newstypes.NewsResponse, error) {
	ctx := context.Background()
	key := "news:" + topic

	data, err := Rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("news not found for topic: %s", topic)
		}
		return nil, fmt.Errorf("failed to get news from redis: %w", err)
	}

	var newsResponse newstypes.NewsResponse
	if err := json.Unmarshal([]byte(data), &newsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal news data: %w", err)
	}

	return &newsResponse, nil
}

func GetUserHistory(userID int64) ([]string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("user:%d:news_history", userID)
	topics, rdbErr := Rdb.LRange(ctx, key, 0, -1).Result()
	if rdbErr != nil {
		return nil, rdbErr
	}
	var values []string

	for _, topic := range topics {
		news_key := fmt.Sprintf("news:%s", topic)
		value, getErr := Rdb.Get(ctx, news_key).Result()
		if getErr != nil {
			log.Printf("Error getting news key %s: %v", news_key, getErr)
			continue
		}
		values = append(values, value)
	}
	return values, nil
}
