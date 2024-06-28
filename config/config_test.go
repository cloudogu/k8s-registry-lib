package config

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
	"testing"
)

func TestCreateConfig(t *testing.T) {
	tests := []struct {
		name string
		data Entries
	}{
		{"empty data", Entries{}},
		{"non-empty data", Entries{"key1": "value1", "key2": "value2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := CreateConfig(tt.data)
			if len(cfg.entries) != len(tt.data) {
				t.Errorf("expected data length %d, got %d", len(tt.data), len(cfg.entries))
			}
			if len(cfg.changeHistory) != 0 {
				t.Errorf("expected change history length 0, got %d", len(cfg.changeHistory))
			}
		})
	}
}

func TestConfig_Set(t *testing.T) {
	cfg := CreateConfig(Entries{
		"key1":             "value1",
		"key2/key21":       "value2",
		"key3/key31/key32": "value3",
	})

	tests := []struct {
		key   Key
		value Value
		xErr  bool
	}{
		{"", "valueErr", true},
		{"/", "valueErr", true},
		{"key1", "newValue1", false},
		{"key1/new", "newValue1", true},
		{"key1/new/new2", "newValue1", true},
		{"key1/new/new3", "newValue1", true},
		{"key1/new/new3/new4", "newValue1", true},
		{"key2", "newValue2", true},
		{"key2/", "newValue2", true},
		{"key2/key21", "newValue2", false},
		{"key2/key21/", "newValue2", true},
		{"key2/key21/new", "newValue2", true},
		{"key2/key21/new/new2", "newValue2", true},
		{"key3", "newValue3", true},
		{"key3/key31", "newValue3", true},
		{"key4", "value4", false},
	}

	for _, tt := range tests {
		t.Run(tt.key.String(), func(t *testing.T) {
			err := cfg.Set(tt.key, tt.value)
			assert.Equal(t, tt.xErr, err != nil)

			if tt.xErr {
				if v, ok := cfg.entries[tt.key]; ok && v == tt.value {
					t.Errorf("new Value for %s Key written, but error has occured", tt.key)
				}

				return
			}

			if v, ok := cfg.entries[tt.key]; !ok || v != tt.value {
				t.Errorf("expected %s for Key %s, got %s", tt.value, tt.key, v)
			}

			lastChange := cfg.changeHistory[len(cfg.changeHistory)-1]
			if lastChange.KeyPath != tt.key || lastChange.Deleted {
				t.Errorf("unexpected change history entry %+v", lastChange)
			}
		})
	}
}

func TestConfig_Exists(t *testing.T) {
	cfg := CreateConfig(Entries{"key1": "value1"})
	tests := []struct {
		key      Key
		expected bool
	}{
		{"key1", true},
		{"key2", false},
	}

	for _, tt := range tests {
		t.Run(tt.key.String(), func(t *testing.T) {
			if exists := cfg.Exists(tt.key); exists != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, exists)
			}
		})
	}
}

func TestConfig_Get(t *testing.T) {
	cfg := CreateConfig(Entries{"key1": "value1"})
	tests := []struct {
		key       Key
		expected  Value
		expectErr bool
	}{
		{"/key1", "value1", false},
		{"key1", "value1", false},
		{"/key2", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.key.String(), func(t *testing.T) {
			value, err := cfg.Get(tt.key)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error %v, got %v", tt.expectErr, err)
			}
			if value != tt.expected {
				t.Errorf("expected Value %s, got %s", tt.expected, value)
			}
		})
	}
}

func TestConfig_GetAll(t *testing.T) {
	data := Entries{"key1": "value1", "key2": "value2"}
	cfg := CreateConfig(data)
	got := cfg.GetAll()
	for k, v := range data {
		if got[k] != v {
			t.Errorf("expected %s for Key %s, got %s", v, k, got[k])
		}
	}
	if len(got) != len(data) {
		t.Errorf("expected length %d, got %d", len(data), len(got))
	}
}

func TestConfig_GetChangeHistory(t *testing.T) {
	changes := []Change{
		{
			KeyPath: "key1",
			Deleted: false,
		},
		{
			KeyPath: "key11/key2/key3",
			Deleted: true,
		},
	}
	cfg := Config{
		entries: map[Key]Value{
			"key1":            "newValue1",
			"key11/key2/key3": "newValue3",
		},
		changeHistory: changes,
	}

	cHistory := cfg.GetChangeHistory()
	assert.Equal(t, changes, cHistory)
}

func TestConfig_Delete(t *testing.T) {
	data := Entries{"key1": "value1", "key2": "value2", "key4": "value4"}
	cfg := CreateConfig(data)
	tests := []struct {
		key      Key
		expected error
	}{
		{"/key1", nil},
		{"/key3", nil},
		{"/Key", nil},
		{"key4", nil},
	}

	for _, tt := range tests {
		t.Run(tt.key.String(), func(t *testing.T) {
			cfg.Delete(tt.key)
			_, ok := cfg.entries[tt.key]
			assert.False(t, ok)
		})
	}
}

func TestConfig_DeleteRecursive(t *testing.T) {
	data := Entries{
		"key0":             "value1",
		"key1/subkey":      "subvalue",
		"key1/subkey2":     "subvalue2",
		"key2":             "value2",
		"key3/key2/key1":   "value3",
		"key3/key2/key11":  "value4",
		"key3/key22/key11": "value5",
	}

	tests := []struct {
		keyToDelete Key
		keysDeleted int
	}{
		{"/key0", 1},
		{"/Key", 0},
		{"/key1/subkey", 1},
		{"key1/", 2},
		{"/key1", 2},
		{"key2", 1},
		{"/key3/key2/", 2},
		{"key3/", 3},
		{"", 7},
	}

	for _, tc := range tests {
		t.Run(tc.keyToDelete.String(), func(t *testing.T) {
			cfg := CreateConfig(maps.Clone(data))
			l := len(cfg.entries)

			cfg.DeleteRecursive(tc.keyToDelete)

			if _, ok := cfg.entries[tc.keyToDelete]; ok {
				t.Error("expected /key1 to be deleted")
			}

			if diff := l - len(cfg.entries); diff != tc.keysDeleted {
				t.Errorf("expected length of config to be %d, got: %d", l-tc.keysDeleted, len(cfg.entries))
			}
		})
	}
}

func TestConfig_DeleteAll(t *testing.T) {
	data := Entries{"key1": "value1", "key2": "value2"}
	cfg := CreateConfig(data)
	cfg.DeleteAll()

	if len(cfg.entries) != 0 {
		t.Errorf("expected all keys to be deleted, got %d keys", len(cfg.entries))
	}
}

func TestCreateGlobalConfig(t *testing.T) {
	cfg := CreateConfig(Entries{"key1": "value1"})
	globalCfg := CreateGlobalConfig(cfg)

	if len(globalCfg.entries) != len(cfg.entries) {
		t.Errorf("expected data length %d, got %d", len(cfg.entries), len(globalCfg.entries))
	}
}

func TestCreateDoguConfig(t *testing.T) {
	cfg := CreateConfig(Entries{"key1": "value1"})
	doguCfg := CreateDoguConfig(cfg)

	if len(doguCfg.entries) != len(cfg.entries) {
		t.Errorf("expected data length %d, got %d", len(cfg.entries), len(doguCfg.entries))
	}
}
