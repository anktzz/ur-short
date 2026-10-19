package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anktzz/ur-short/internal/store"
)

type fakeStore struct {
	urls map[uint64]string
	next uint64
	err  error
}

func newFake() *fakeStore {
	return &fakeStore{urls: map[uint64]string{}, next: 1}
}

func (f *fakeStore) Create(ctx context.Context, longURL string) (uint64, error) {
	if f.err != nil {
		return 0, f.err
	}
	id := f.next
	f.next++
	f.urls[id] = longURL
	return id, nil
}

func (f *fakeStore) Get(ctx context.Context, id uint64) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	long, ok := f.urls[id]
	if !ok {
		return "", store.ErrNotFound
	}
	return long, nil
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestShorten(t *testing.T) {
	h := New(newFake(), "http://test").Routes()
	rec := do(h, http.MethodPost, "/api/shorten", `{"long_url":"https://example.com/a"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"short_url":"http://test/1"`) {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestShortenInvalid(t *testing.T) {
	h := New(newFake(), "http://test").Routes()
	bodies := []string{
		`not json`,
		`{"long_url":""}`,
		`{"long_url":"example.com"}`,
		`{"long_url":"ftp://example.com"}`,
	}
	for _, b := range bodies {
		rec := do(h, http.MethodPost, "/api/shorten", b)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %q: status = %d, want 400", b, rec.Code)
		}
	}
}

func TestRedirect(t *testing.T) {
	f := newFake()
	f.urls[1] = "https://example.com/a"
	h := New(f, "http://test").Routes()
	rec := do(h, http.MethodGet, "/1", "")
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "https://example.com/a" {
		t.Errorf("Location = %q", loc)
	}
}

func TestRedirectNotFound(t *testing.T) {
	h := New(newFake(), "http://test").Routes()
	for _, path := range []string{"/zz", "/a-b"} {
		rec := do(h, http.MethodGet, path, "")
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: status = %d, want 404", path, rec.Code)
		}
	}
}

func TestStoreError(t *testing.T) {
	f := newFake()
	f.err = errors.New("boom")
	h := New(f, "http://test").Routes()

	rec := do(h, http.MethodPost, "/api/shorten", `{"long_url":"https://example.com/a"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("shorten: status = %d, want 500", rec.Code)
	}
	rec = do(h, http.MethodGet, "/1", "")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("redirect: status = %d, want 500", rec.Code)
	}
}
