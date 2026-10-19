package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/anktzz/ur-short/internal/handler"
	"github.com/anktzz/ur-short/internal/store"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	port := getenv("PORT", "8080")
	baseURL := getenv("BASE_URL", "http://localhost:"+port)

	db, err := store.NewPostgres(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{Addr: getenv("REDIS_ADDR", "localhost:6379")})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Println("redis unavailable:", err)
	}

	cached := store.NewCached(db, rdb, 24*time.Hour)
	h := handler.New(cached, baseURL)
	log.Println("listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, h.Routes()))
}
