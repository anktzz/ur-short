package main

import (
	"context"
	"log"
	"net/http"
	"os"

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
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	port := getenv("PORT", "8080")
	baseURL := getenv("BASE_URL", "http://localhost:"+port)

	db, err := store.NewPostgres(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	h := handler.New(db, baseURL)
	log.Println("listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, h.Routes()))
}
