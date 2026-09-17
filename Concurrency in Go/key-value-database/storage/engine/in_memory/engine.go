package in_memory

import (
	"context"
	"log/slog"
)

type Engine struct {
	table  *HashTable
	logger *slog.Logger
}

func NewEngine(logger *slog.Logger) *Engine {
	return &Engine{
		logger: logger,
		table:  NewHashTable(),
	}
}

func (e *Engine) Set(ctx context.Context, key string, val string) {
	e.table.Set(key, val)
	e.logger.DebugContext(ctx, "successfully set")
}

func (e *Engine) Get(ctx context.Context, key string) (string, bool) {
	return e.table.Get(key)
}

func (e *Engine) Del(ctx context.Context, key string) {
	e.table.Del(key)
}
