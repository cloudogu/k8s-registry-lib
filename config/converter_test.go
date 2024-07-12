package config

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestMapToConfig(t *testing.T) {
	testCases := []struct {
		name        string
		sourceMap   map[string]any
		expected    Entries
		expectFail  bool
		expectedErr string
	}{
		{
			name: "Simple map conversion",
			sourceMap: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			expected: Entries{
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
					"child3": "123",
				},
			},
			expected: Entries{
				"parent/child1": "value1",
				"parent/child2": "value2",
				"parent/child3": "123",
			},
			expectFail: false,
		},
		{
			name: "invalid yaml",
			sourceMap: map[string]any{
				"parent": map[string]any{
					"child1": &Entries{},
				},
			},
			expected:    Entries{},
			expectFail:  true,
			expectedErr: "could not convert &map[] to value (string)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result Entries

			err := mapToConfig(tc.sourceMap, &result, "")
			if tc.expectFail {
				assert.Error(t, err)
				if tc.expectedErr != "" {
					assert.Equal(t, err.Error(), "could not convert &map[] to value (string)")
				}
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Unexpected result. Got: %v, Expected: %v", result, tc.expected)
			}
		})
	}
}

func TestConfigToMap(t *testing.T) {
	testCases := []struct {
		name       string
		sourceData Entries
		prefix     string
		expected   map[string]any
	}{
		{
			name: "Simple map conversion",
			sourceData: Entries{
				"key1": "value1",
				"key2": "value2",
			},
			prefix:   "",
			expected: map[string]any{"key1": "value1", "key2": "value2"},
		},
		{
			name: "Nested map conversion",
			sourceData: Entries{
				"parent/child1": "value1",
				"parent/child2": "value2",
			},
			prefix:   "",
			expected: map[string]any{"parent": map[string]any{"child1": "value1", "child2": "value2"}},
		},
		{
			name: "Complex nested map conversion",
			sourceData: Entries{
				"grandparent/parent/child1": "value1",
				"grandparent/parent/child2": "value2",
			},
			prefix:   "",
			expected: map[string]any{"grandparent": map[string]any{"parent": map[string]any{"child1": "value1", "child2": "value2"}}},
		},
		{
			name: "Conversion with prefix",
			sourceData: Entries{
				"parent/child1": "value1",
				"parent/child2": "value2",
			},
			prefix:   "parent/",
			expected: map[string]any{"child1": "value1", "child2": "value2"},
		},
		{
			name: "Prefix not found",
			sourceData: Entries{
				"parent/child1": "value1",
				"parent/child2": "value2",
			},
			prefix:   "nonexistent/",
			expected: map[string]any{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := configToMap(tc.sourceData, tc.prefix)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Unexpected result. Got: %v, Expected: %v", result, tc.expected)
			}
		})
	}
}

func TestYamlConverter_Read(t *testing.T) {
	testCases := []struct {
		name       string
		yamlInput  string
		nilReader  bool
		expected   Entries
		expectFail bool
	}{
		{
			name: "Simple YAML",
			yamlInput: `
key1: value1
key2: value2
`,
			expected: Entries{
				"key1": "value1",
				"key2": "value2",
			},
			expectFail: false,
		},
		{
			name: "Nested YAML",
			yamlInput: `
parent:
 child1: value1
 child2: value2
`,
			expected: Entries{
				"parent/child1": "value1",
				"parent/child2": "value2",
			},
			expectFail: false,
		},
		{
			name:       "Empty YAML",
			yamlInput:  ``,
			expected:   Entries{},
			expectFail: true,
		},
		{
			name: "Nil Reader",
			yamlInput: `
parent:
 child1: "123"
`,
			nilReader:  true,
			expectFail: true,
		},
		{
			name: "invalid yaml",
			yamlInput: `
parent:
child1; 123
`,
			expected:   Entries{},
			expectFail: true,
		},
		{
			name: "invalid yaml",
			yamlInput: `
parent:
 child1: 123
`,
			expected:   Entries{},
			expectFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var reader io.Reader
			if tc.nilReader {
				reader = nil
			} else {
				reader = strings.NewReader(tc.yamlInput)
			}

			yc := &YamlConverter{}
			result, err := yc.Read(reader)
			if tc.expectFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestYamlConverter_Write(t *testing.T) {
	testCases := []struct {
		name     string
		data     Entries
		expected string
	}{
		{
			name: "Simple Entries",
			data: Entries{
				"key1": "value1",
				"key2": "value2",
			},
			expected: "key1: value1\nkey2: value2\n",
		},
		{
			name: "Nested Entries",
			data: Entries{
				"parent/child1": "value1",
				"parent/child2": "value2",
			},
			expected: "parent:\n    child1: value1\n    child2: value2\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buffer bytes.Buffer
			yc := &YamlConverter{}
			err := yc.Write(&buffer, tc.data)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, buffer.String())
		})
	}
}
