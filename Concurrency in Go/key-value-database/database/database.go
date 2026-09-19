package database

import (
	"context"
	"errors"
	"fmt"
	"key-value-database/database/compute"
	"key-value-database/database/storage"

	"log/slog"
)

type computeLayer interface {
	Parse(ctx context.Context, raw string) (compute.Query, error)
}

type storageLayer interface {
	Set(context.Context, string, string) error
	Get(context.Context, string) (string, error)
	Del(context.Context, string) error
}

type Database struct {
	computeLayer computeLayer
	storageLayer storageLayer
	logger       *slog.Logger
}

func NewDatabase(computeLayer computeLayer, storageLayer storageLayer, logger *slog.Logger) (*Database, error) {
	if computeLayer == nil {
		return nil, errors.New("compute layer not provided")
	}

	if storageLayer == nil {
		return nil, errors.New("storagev layer not provifrf")
	}

	if logger == nil {
		return nil, errors.New("logger not provided")
	}

	return &Database{
		computeLayer: computeLayer,
		storageLayer: storageLayer,
		logger:       logger,
	}, nil
}

func (d *Database) HandleQuery(ctx context.Context, queryStr string) string {
	d.logger.DebugContext(ctx, "handling query...", "query", queryStr)

	query, err := d.computeLayer.Parse(ctx, queryStr)
	if err != nil {
		return fmt.Sprintf("[error] %s", err.Error())
	}

	switch query.CommandID() {
	case compute.GET:
		return d.handleGetQuery(ctx, query)
	case compute.SET:
		return d.handleSetQuery(ctx, query)
	case compute.DEL:
		return d.handleDelQuery(ctx, query)
	}

	d.logger.Error(
		"compute layer is incorrect",
		"command_id", query.CommandID(),
	)

	return "[error] internal error"
}

func (d *Database) handleGetQuery(ctx context.Context, query compute.Query) string {
	arguments := query.Arguments()
	result, err := d.storageLayer.Get(ctx, arguments[0])
	if err == storage.ErrNotFound {
		return "[not found]"
	} else if err != nil {
		return fmt.Sprintf("[error] %s", err.Error())
	}

	return fmt.Sprintf("[ok] %s", result)
}

func (d *Database) handleSetQuery(ctx context.Context, query compute.Query) string {
	arguments := query.Arguments()
	if err := d.storageLayer.Set(ctx, arguments[0], arguments[1]); err != nil {
		return fmt.Sprintf("[error] %s", err.Error())
	}

	return "[ok]"
}

func (d *Database) handleDelQuery(ctx context.Context, query compute.Query) string {
	arguments := query.Arguments()
	if err := d.storageLayer.Del(ctx, arguments[0]); err != nil {
		return fmt.Sprintf("[error] %s", err.Error())
	}

	return "[ok]"
}
