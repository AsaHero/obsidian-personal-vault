package compute

import (
	"context"
	"errors"
	"log/slog"
	"strings"
)

var (
	ErrInvalidQuery           = errors.New("invalid query format")
	ErrInvalidCommand         = errors.New("invalid command name")
	ErrInvalidNumberArguments = errors.New("invalid number of arguments")
)

type parser struct {
	logger *slog.Logger
}

func NewParser(logger *slog.Logger) *parser {
	return &parser{
		logger: logger,
	}
}

func (e *parser) Parse(ctx context.Context, raw string) (Query, error) {
	tokens := strings.Fields(raw)

	if len(tokens) < 1 {
		return Query{}, ErrInvalidQuery
	}

	q := Query{}
	cmd := tokens[0]

	switch cmd {
	case getCommand:
		if len(tokens) < 2 {
			return Query{}, ErrInvalidNumberArguments
		}

		q.command = GET
		q.arguments = tokens[1:]
	case setCommand:
		if len(tokens) < 3 {
			return Query{}, ErrInvalidNumberArguments
		}

		q.command = SET
		q.arguments = tokens[1:]
	case delCommand:
		if len(tokens) < 2 {
			return Query{}, ErrInvalidNumberArguments
		}

		q.command = DEL
		q.arguments = tokens[1:]
	default:
		return Query{}, ErrInvalidCommand
	}

	return q, nil
}
