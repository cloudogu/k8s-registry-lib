package config

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestMapToConfig(t *testing.T) {
	testCases := []struct {
		name       string
		sourceMap  map[string]any
		expected   Data
		expectFail bool
	}{
		{
			name: "Simple map conversion",
			sourceMap: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			expected: Data{
				"key1": "value1",
				"key2": "value2",
			},
			expectFail: false,
		},
		{
			name: "Nested map conversion",
			sourceMap: map[string]any{
				"parent": map[string]any{
					"child1": "value1",
					"child2": "value2",
				},
			},
			expected: Data{
				"parent/child1": "value1",
				"parent/child2": "value2",
			},
			expectFail: false,
		},
		{
			name: "invalid values",
			sourceMap: map[string]any{
				"parent": map[string]any{
					"child1": 123,
				},
			},
			expectFail: true,
			expected:   map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result Data

			err := mapToConfig(tc.sourceMap, &result, "")
			if tc.expectFail {
				assert.Contains(t, err.Error(), "could not convert 123 to string")
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Unexpected result. Got: %v, Expected: %v", result, tc.expected)
			}
		})
	}
}
