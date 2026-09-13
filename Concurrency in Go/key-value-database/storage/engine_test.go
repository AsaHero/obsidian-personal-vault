package storage_test

import (
	"key-value-database/storage"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineSet(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		engine storage.Engine
		key    string
		value  string
	}{
		"set with single string value": {
			engine: storage.NewNegine(),
			key:    "key",
			value:  "hello",
		},
		"set with multiple string value": {
			engine: storage.NewNegine(),
			key:    "key",
			value:  "hello world",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			test.engine.Set(test.key, test.value)

			value, err := test.engine.Get(test.key)
			require.NoError(t, err)
			assert.Equal(t, value, test.value)
		})
	}
}
