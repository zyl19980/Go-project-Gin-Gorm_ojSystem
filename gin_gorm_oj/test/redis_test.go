package test

import (
	"context"
	"fmt"
	"gin_gorm_oj/models"
	"github.com/go-redis/redis/v8"
	"testing"
	"time"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

func TestRedisSet(t *testing.T) {
	rdb.Set(ctx, "name", "mmc", time.Second*10)
}

func TestRedisGet(t *testing.T) {
	s, err := rdb.Get(ctx, "name").Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println(s)
}

func TestRedisGetByModel(t *testing.T) {
	s, err := models.RDB.Get(ctx, "name").Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println(s)
}
