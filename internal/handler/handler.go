package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/anktzz/ur-short/internal/base62"
	"github.com/anktzz/ur-short/internal/store"
)

type Store interface {
	Create(ctx context.Context, longURL string) (uint64, error)
	Get(ctx context.Context, id uint64) (string, error)
}

type Handler struct {
	store   Store
	baseURL string
}

func New(s Store, baseURL string) *Handler {
	return &Handler{store: s, baseURL: baseURL}
}

type shortenRequest struct {
	LongURL string `json:"long_url"`
}

type shortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	u, err := url.ParseRequestURI(req.LongURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	id, err := h.store.Create(r.Context(), req.LongURL)
	if err != nil {
		log.Println("create failed:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	code := base62.Encode(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(shortenResponse{
		ShortCode: code,
		ShortURL:  h.baseURL + "/" + code,
	})
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	id, err := base62.Decode(code)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	long, err := h.store.Get(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Println("get failed:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, long, http.StatusFound)
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/api/shorten", h.Shorten)
	r.Get("/{code}", h.Redirect)
	return r
}
