package main

import (
	"context"

	"github.com/go-redis/redis/v9"
)

var ctx = context.Background()

// Creates and returns a redis client which will be used in future operations
func getRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return rdb
}

func storeMemory(rdb *redis.Client, key string, value string) error {
	// Store "value" under the "key" in redis
	err := rdb.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func fetchMemory(rdb *redis.Client, key string) (*MemoryStruct, error) {
	// Get the value stored under the "key" in Redis, return {key, value} struct
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	res := &MemoryStruct{}
	res.Key = key
	res.Value = val
	return res, nil
}
