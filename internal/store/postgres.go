package store

import (
	"context"
	"errors"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) Create(ctx context.Context, longURL string) (uint64, error) {
	var id int64
	err := p.pool.QueryRow(ctx, "INSERT INTO urls (long_url) VALUES ($1) RETURNING id", longURL).Scan(&id)
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (p *Postgres) Get(ctx context.Context, id uint64) (string, error) {
	if id > math.MaxInt64 {
		return "", ErrNotFound
	}
	var long string
	err := p.pool.QueryRow(ctx, "SELECT long_url FROM urls WHERE id = $1", id).Scan(&long)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return long, err
}
