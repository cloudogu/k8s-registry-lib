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
			nCfg, err := cfg.Set(tt.key, tt.value)
			assert.Equal(t, tt.xErr, err != nil)

			if tt.xErr {
				if v, ok := nCfg.entries[tt.key]; ok && v == tt.value {
					t.Errorf("new Value for %s Key written, but error has occured", tt.key)
				}

				return
			}

			if v, ok := nCfg.entries[tt.key]; !ok || v != tt.value {
				t.Errorf("expected %s for Key %s, got %s", tt.value, tt.key, v)
			}

			lastChange := nCfg.changeHistory[len(nCfg.changeHistory)-1]
			if lastChange.KeyPath != tt.key || lastChange.Deleted {
				t.Errorf("unexpected change history entry %+v", lastChange)
			}
		})
	}
}

func TestConfig_Get(t *testing.T) {
	cfg := CreateConfig(Entries{"key1": "value1"})
	tests := []struct {
		key      Key
		expected Value
		exists   bool
	}{
		{"/key1", "value1", true},
		{"key1", "value1", true},
		{"/key2", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.key.String(), func(t *testing.T) {
			value, ok := cfg.Get(tt.key)
			assert.Equal(t, tt.exists, ok)

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

func TestConfig_Diff(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		oCfg    Config
		expMods []DiffResult
	}{
		{
			name: "Same config values",
			cfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
					"k2": "v2",
					"k3": "v3",
				},
			},
			oCfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
					"k2": "v2",
					"k3": "v3",
				},
			},
			expMods: make([]DiffResult, 0),
		},
		{
			name: "Other differs in k2 and k3",
			cfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
					"k2": "v2",
					"k3": "v3",
				},
			},
			oCfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
					"k2": "v3",
					"k3": "v4",
				},
			},
			expMods: []DiffResult{
				{
					Key:        "k2",
					Value:      OptionalValue{String: "v2", Exists: true},
					OtherValue: OptionalValue{String: "v3", Exists: true},
				},
				{
					Key:        "k3",
					Value:      OptionalValue{String: "v3", Exists: true},
					OtherValue: OptionalValue{String: "v4", Exists: true},
				},
			},
		},
		{
			name: "Missing key k1 in config",
			cfg: Config{
				entries: map[Key]Value{
					"k2": "v2",
					"k3": "v3",
				},
			},
			oCfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
					"k2": "v2",
					"k3": "v3",
				},
			},
			expMods: []DiffResult{
				{
					Key:        "k1",
					Value:      OptionalValue{Exists: false},
					OtherValue: OptionalValue{String: "v1", Exists: true},
				},
			},
		},
		{
			name: "Missing key k2 in other config",
			cfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
					"k2": "v2",
					"k3": "v3",
				},
			},
			oCfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
					"k3": "v3",
				},
			},
			expMods: []DiffResult{
				{
					Key:        "k2",
					Value:      OptionalValue{String: "v2", Exists: true},
					OtherValue: OptionalValue{Exists: false},
				},
			},
		},
		{
			name: "Missing key k1 in config and empty key in other config",
			cfg: Config{
				entries: map[Key]Value{
					"k2": "v2",
					"k3": "v3",
				},
			},
			oCfg: Config{
				entries: map[Key]Value{
					"k1": "",
					"k2": "v2",
					"k3": "v3",
				},
			},
			expMods: []DiffResult{
				{
					Key:        "k1",
					Value:      OptionalValue{Exists: false},
					OtherValue: OptionalValue{String: "", Exists: true},
				},
			},
		},
		{
			name: "Empty key k2 in config and missing key in other config",
			cfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
					"k2": "",
					"k3": "v3",
				},
			},
			oCfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
					"k3": "v3",
				},
			},
			expMods: []DiffResult{
				{
					Key:        "k2",
					Value:      OptionalValue{String: "", Exists: true},
					OtherValue: OptionalValue{Exists: false},
				},
			},
		},
		{
			name: "Empty key k2 in config and empty key k2 in other config",
			cfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
					"k2": "",
					"k3": "v3",
				},
			},
			oCfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
					"k2": "",
					"k3": "v3",
				},
			},
			expMods: []DiffResult{},
		},
		{
			name: "Missing key k1 in config and missing key k1 in other config",
			cfg: Config{
				entries: map[Key]Value{
					"k2": "v2",
					"k3": "v3",
				},
			},
			oCfg: Config{
				entries: map[Key]Value{
					"k2": "v2",
					"k3": "v3",
				},
			},
			expMods: []DiffResult{},
		},
		{
			name: "Multiple keys keys added to other config",
			cfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
				},
			},
			oCfg: Config{
				entries: map[Key]Value{
					"k1": "new",
					"k2": "v2",
					"k3": "v3",
					"k4": "v4",
					"k5": "v5",
				},
			},
			expMods: []DiffResult{
				{
					Key:        "k1",
					Value:      OptionalValue{String: "v1", Exists: true},
					OtherValue: OptionalValue{String: "new", Exists: true},
				},
				{
					Key:        "k2",
					Value:      OptionalValue{Exists: false},
					OtherValue: OptionalValue{String: "v2", Exists: true},
				},
				{
					Key:        "k3",
					Value:      OptionalValue{Exists: false},
					OtherValue: OptionalValue{String: "v3", Exists: true},
				},
				{
					Key:        "k4",
					Value:      OptionalValue{Exists: false},
					OtherValue: OptionalValue{String: "v4", Exists: true},
				},
				{
					Key:        "k5",
					Value:      OptionalValue{Exists: false},
					OtherValue: OptionalValue{String: "v5", Exists: true},
				},
			},
		},
		{
			name: "Compare with empty config",
			cfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
				},
			},
			oCfg: Config{
				entries: make(Entries),
			},
			expMods: []DiffResult{
				{
					Key:        "k1",
					Value:      OptionalValue{String: "v1", Exists: true},
					OtherValue: OptionalValue{Exists: false},
				},
			},
		},
		{
			name: "Compare with nil config",
			cfg: Config{
				entries: map[Key]Value{
					"k1": "v1",
				},
			},
			oCfg: Config{},
			expMods: []DiffResult{
				{
					Key:        "k1",
					Value:      OptionalValue{String: "v1", Exists: true},
					OtherValue: OptionalValue{Exists: false},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mods := tc.cfg.Diff(tc.oCfg)

			assert.Equal(t, len(tc.expMods), len(mods))

			for _, xm := range tc.expMods {
				found := false

				for _, m := range mods {
					if xm.Key == m.Key {
						found = true
						assert.Equal(t, xm, m)
						continue
					}
				}

				if !found {
					assert.Fail(t, "expected modifications and actual modifications differs", "expected", tc.expMods, "actual", mods)
				}
			}
		})
	}
}
