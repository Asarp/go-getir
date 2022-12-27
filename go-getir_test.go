package main

import (
	"testing"
	"time"
)

func TestFetchDBValid(t *testing.T) {
	DbReponse, _ := fetchDB(time.Date(0, time.August, 12, 12, 12, 12, 12, time.UTC), time.Now(), 0, 10000000000)
	if DbReponse.Code != 0 {
		t.Errorf("Code was incorrect, got: %d, want: %d.", DbReponse.Code, 0)
	}
}

func TestStoreMemoryValid(t *testing.T) {
	redisClient := getRedisClient()
	if err := storeMemory(redisClient, "test", "valid"); err != nil {
		t.Errorf("Error occured when writing to Redis")
	}
}

func TestFetchMemoryValid(t *testing.T) {
	redisClient := getRedisClient()
	storeMemory(redisClient, "test", "valid")
	redisResponse, err := fetchMemory(redisClient, "test")
	if err != nil || redisResponse.Key != "test" || redisResponse.Value != "valid" {
		t.Errorf("Error occured when fetching from Redis. The expected response for the key 'test' was 'value', got: %s'", redisResponse.Value)
	}
}
