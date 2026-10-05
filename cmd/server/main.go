package main

import (
	"log"
	"net/http"

	"github.com/anktzz/ur-short/internal/handler"
)

func main() {
	h := handler.New()
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", h.Routes()))
}
