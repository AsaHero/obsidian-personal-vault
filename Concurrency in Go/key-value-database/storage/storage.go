package storage

import (
	"context"
	"errors"
	"log/slog"
)

var (
	ErrNotFound = errors.New("not found")
)

type Engine interface {
	Set(ctx context.Context, key string, value string)
	Get(ctx context.Context, key string) (string, bool)
	Del(ctx context.Context, key string)
}

type Storage struct {
	engine Engine
	logger *slog.Logger
}

func NewStorage(engine Engine, logger *slog.Logger) (*Storage, error) {
	if engine == nil {
		return nil, errors.New("engine is not provided")
	}

	if logger == nil {
		return nil, errors.New("logger is not provided")
	}

	return &Storage{
		engine: engine,
		logger: logger,
	}, nil
}

func (s *Storage) Set(ctx context.Context, key string, value string) error {
	s.engine.Set(ctx, key, value)
	return nil
}

func (s *Storage) Get(ctx context.Context, key string) (string, error) {
	value, found := s.engine.Get(ctx, key)
	if !found {
		return "", ErrNotFound
	}

	return value, nil
}

func (s *Storage) Del(ctx context.Context, key string) error {
	s.engine.Del(ctx, key)
	return nil
}
