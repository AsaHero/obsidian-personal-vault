package storage

import "errors"

var (
	ErrNotFound = errors.New("not found")
)

type Engine interface {
	Set(key string, val string)
	Get(key string) (val string, err error)
	Del(key string)
}

type engine struct {
	db map[string]string
}

func NewNegine() Engine {
	return &engine{
		db: make(map[string]string),
	}
}

func (e *engine) Set(key string, val string) {
	e.db[key] = val
}

func (e *engine) Get(key string) (val string, err error) {
	val, ok := e.db[key]
	if !ok {
		return "", ErrNotFound
	}

	return val, nil
}

func (e *engine) Del(key string) {
	delete(e.db, key)
}
