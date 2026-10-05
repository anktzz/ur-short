package handler 

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/anktzz/ur-short/internal/base62"
)

type Handler struct {
	mu sync.Mutex
	urls map[uint64]string
	next uint64
}

func New() *Handler {
	return &Handler{urls: make(map[uint64]string), next: 1}
}

type shortenRequest struct {
	LongUrl string `json:"long_url"`
}

type shortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortUrl string `json:"short_url"`
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {

	var req shortenRequest
	if err:= json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Json", http.StatusBadRequest)
		return
	}

	u, err := url.ParseRequestURI(req.LongUrl) 
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		http.Error(w, "Invalid Url", http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	id := h.next
	h.next++
	h.urls[id] = req.LongUrl
	h.mu.Unlock()

	code := base62.Encode(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(shortenResponse{
		ShortCode: code,
		ShortUrl: "http://localhost:8080" + code,
	})
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	id, err := base62.Decode(code)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	h.mu.Lock()
	long, ok := h.urls[id]
	h.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
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
