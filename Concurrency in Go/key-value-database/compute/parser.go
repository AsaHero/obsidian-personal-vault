package compute

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidQuery    = errors.New("invalid query")
	ErrInvalidCommand  = errors.New("invalid command")
	ErrInvalidArgument = errors.New("invalid arguments")
)

var (
	argumentRegex = regexp.MustCompile(`\w+`)
)

type Command uint8

const (
	GET Command = iota
	SET
	DEL
)

type Query struct {
	Value string
	Key   string
	Cmd   Command
}

type Parser interface {
	Parse() (*Query, error)
}

type parser struct {
	exp string
}

func NewParser(exp string) Parser {
	return &parser{
		exp: exp,
	}
}

func (e *parser) Parse() (*Query, error) {
	splits := strings.Split(e.exp, " ")

	if len(splits) < 1 {
		return nil, ErrInvalidQuery
	}

	q := Query{}

	switch splits[0] {
	case "GET":
		q.Cmd = GET
		q.Key = splits[1]

		if !argumentRegex.MatchString(q.Key) {
			return nil, ErrInvalidArgument
		}

		return &q, nil
	case "SET":
		q.Cmd = SET
		if len(splits) < 3 {
			return nil, ErrInvalidQuery
		}

		q.Key = splits[1]
		if !argumentRegex.MatchString(q.Key) {
			return nil, ErrInvalidArgument
		}

		q.Value = strings.Join(splits[2:], " ")
		if !argumentRegex.MatchString(q.Value) {
			return nil, ErrInvalidArgument
		}

		return &q, nil
	case "DEL":
		q.Cmd = DEL
		q.Key = splits[1]

		if !argumentRegex.MatchString(q.Key) {
			return nil, ErrInvalidArgument
		}

		return &q, nil
	default:
		return nil, ErrInvalidCommand
	}
}
