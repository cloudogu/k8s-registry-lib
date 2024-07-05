package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKeyFilter(t *testing.T) {
	diffs := []DiffResult{
		{Key: "key1"},
		{Key: "key2/key3"},
		{Key: "key4/key5/key6"},
		{Key: "key4/key5/key7"},
	}

	tests := []struct {
		key     Key
		xResult bool
	}{
		{"key1", true},
		{"key2", false},
		{"key3", false},
		{"key4", false},
		{"key2/key3", true},
		{"key4/key5", false},
		{"key4/key5/key6", true},
		{"key4/key5/key7", true},
		{"key4/key5/key8", false},
	}

	for _, tc := range tests {
		t.Run(string(tc.key), func(t *testing.T) {
			filter := KeyFilter(tc.key)
			assert.Equal(t, tc.xResult, filter(diffs))
		})
	}
}

func TestDirectoryFilter(t *testing.T) {
	diffs := []DiffResult{
		{Key: "key1"},
		{Key: "key2/key3"},
		{Key: "key4/key5/key6"},
		{Key: "key4/key5/key7"},
	}

	tests := []struct {
		key     Key
		xResult bool
	}{
		{"", true},
		{"/", true},
		{"key1", false},
		{"key2", true},
		{"key3", false},
		{"key4", true},
		{"key2/key3", false},
		{"key4/key5", true},
		{"key4/key5/key6", false},
		{"key4/key5/key7", false},
		{"key4/key5/key8", false},
	}

	for _, tc := range tests {
		t.Run(string(tc.key), func(t *testing.T) {
			filter := DirectoryFilter(tc.key)
			assert.Equal(t, tc.xResult, filter(diffs))
		})
	}
}
