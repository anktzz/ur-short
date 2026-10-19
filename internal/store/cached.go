package store

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store interface {
	Create(ctx context.Context, longURL string) (uint64, error)
	Get(ctx context.Context, id uint64) (string, error)
}

type Cached struct {
	db  Store
	rdb *redis.Client
	ttl time.Duration
}

func NewCached(db Store, rdb *redis.Client, ttl time.Duration) *Cached {
	return &Cached{db: db, rdb: rdb, ttl: ttl}
}

func (c *Cached) Create(ctx context.Context, longURL string) (uint64, error) {
	return c.db.Create(ctx, longURL)
}

func (c *Cached) Get(ctx context.Context, id uint64) (string, error) {
	key := "url:" + strconv.FormatUint(id, 10)

	val, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		log.Println("cache hit", key)
		return val, nil
	}
	if !errors.Is(err, redis.Nil) {
		log.Println("redis get failed:", err)
	}

	log.Println("cache miss", key)
	long, err := c.db.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if err := c.rdb.Set(ctx, key, long, c.ttl).Err(); err != nil {
		log.Println("redis set failed:", err)
	}
	return long, nil
}
