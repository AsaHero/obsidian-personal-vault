package compute_test

import (
	"key-value-database/compute"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserParas(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		exp          string
		query        compute.Query
		expectErr    bool
		expectNilObj bool
	}{
		"invalid query": {
			exp:          "",
			expectErr:    true,
			expectNilObj: true,
		},
		"invalid commad": {
			exp:          "DELETE key value",
			expectErr:    true,
			expectNilObj: true,
		},
		"valid GET": {
			exp: "GET key1",
			query: compute.Query{
				Cmd: compute.GET,
				Key: "key1",
			},
		},
		"invalid GET": {
			exp:          "GET ",
			expectErr:    true,
			expectNilObj: true,
		},
		"valid SET": {
			exp: "SET key1 value1",
			query: compute.Query{
				Cmd:   compute.SET,
				Key:   "key1",
				Value: "value1",
			},
		},
		"invalid SET": {
			exp:          "SET ",
			expectErr:    true,
			expectNilObj: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			query, err := compute.NewParser(test.exp).Parse()
			if test.expectErr {
				require.Error(t, err)
			}

			if !test.expectNilObj {
				assert.NotNil(t, query)
				assert.Equal(t, *query, test.query)
			}

		})
	}
}
