package main

import (
	"context"
	"time"
)

// TODO 9: хранилище на github.com/jackc/pgx/v5/pgxpool.

type Store struct {
	// TODO 9
}

func NewStore(ctx context.Context, url string) (*Store, error) {
	panic("TODO 9: NewStore")
}

func (s *Store) Close() {
	panic("TODO 9: Close")
}

func (s *Store) LoadTags(ctx context.Context) ([]*Tag, error) {
	panic("TODO 9: LoadTags")
}

func (s *Store) SaveSamples(ctx context.Context, ts time.Time, samples []Sample) error {
	panic("TODO 9: SaveSamples")
}

func (s *Store) SaveEvent(ctx context.Context, ts time.Time, t *Tag, st State, v float64, msg string) error {
	panic("TODO 9: SaveEvent")
}
