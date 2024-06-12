package config

import (
	"golang.org/x/exp/maps"
	"testing"
)

func TestCreateConfig(t *testing.T) {
	tests := []struct {
		name string
		data Data
	}{
		{"empty data", Data{}},
		{"non-empty data", Data{"/key1": "value1", "/key2": "value2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := CreateConfig(tt.data)
			if len(cfg.Data) != len(tt.data) {
				t.Errorf("expected data length %d, got %d", len(tt.data), len(cfg.Data))
			}
			if len(cfg.ChangeHistory) != 0 {
				t.Errorf("expected change history length 0, got %d", len(cfg.ChangeHistory))
			}
		})
	}
}

func TestConfig_Set(t *testing.T) {
	cfg := CreateConfig(Data{})
	tests := []struct {
		key   string
		value string
	}{
		{"/key1", "value1"},
		{"/key2", "value2"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			cfg.Set(tt.key, tt.value)
			if v, ok := cfg.Data[tt.key]; !ok || v != tt.value {
				t.Errorf("expected %s for key %s, got %s", tt.value, tt.key, v)
			}
			lastChange := cfg.ChangeHistory[len(cfg.ChangeHistory)-1]
			if lastChange.KeyPath != tt.key || lastChange.Deleted {
				t.Errorf("unexpected change history entry %+v", lastChange)
			}
		})
	}
}

func TestConfig_Exists(t *testing.T) {
	cfg := CreateConfig(Data{"/key1": "value1"})
	tests := []struct {
		key      string
		expected bool
	}{
		{"/key1", true},
		{"/key2", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if exists := cfg.Exists(tt.key); exists != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, exists)
			}
		})
	}
}

func TestConfig_Get(t *testing.T) {
	cfg := CreateConfig(Data{"/key1": "value1"})
	tests := []struct {
		key       string
		expected  string
		expectErr bool
	}{
		{"/key1", "value1", false},
		{"/key2", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			value, err := cfg.Get(tt.key)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error %v, got %v", tt.expectErr, err)
			}
			if value != tt.expected {
				t.Errorf("expected value %s, got %s", tt.expected, value)
			}
		})
	}
}

func TestConfig_GetAll(t *testing.T) {
	data := Data{"/key1": "value1", "/key2": "value2"}
	cfg := CreateConfig(data)
	got := cfg.GetAll()
	for k, v := range data {
		if got[k] != v {
			t.Errorf("expected %s for key %s, got %s", v, k, got[k])
		}
	}
	if len(got) != len(data) {
		t.Errorf("expected length %d, got %d", len(data), len(got))
	}
}

func TestConfig_Delete(t *testing.T) {
	data := Data{"/key1": "value1", "/key2": "value2"}
	cfg := CreateConfig(data)
	tests := []struct {
		key      string
		expected error
	}{
		{"/key1", nil},
		{"/key3", nil},
		{"/key", nil},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			err := cfg.Delete(tt.key)
			if (err != nil) != (tt.expected != nil) {
				t.Errorf("expected error %v, got %v", tt.expected, err)
			}
		})
	}
}

func TestConfig_DeleteRecursive(t *testing.T) {
	data := Data{
		"/key0":             "value1",
		"/key1/subkey":      "subvalue",
		"/key1/subkey2":     "subvalue2",
		"/key2":             "value2",
		"/key3/key2/key1":   "value3",
		"/key3/key2/key11":  "value4",
		"/key3/key22/key11": "value5",
	}

	tests := []struct {
		keyToDelete string
		keysDeleted int
	}{
		{"/key0", 1},
		{"/key", 0},
		{"/key1/subkey", 1},
		{"/key1/", 2},
		{"/key1", 2},
		{"/key2", 1},
		{"/key3/key2/", 2},
		{"/key3/", 3},
	}

	for _, tc := range tests {
		cfg := CreateConfig(maps.Clone(data))
		l := len(cfg.Data)

		cfg.DeleteRecursive(tc.keyToDelete)

		if _, ok := cfg.Data[tc.keyToDelete]; ok {
			t.Error("expected /key1 to be deleted")
		}

		if diff := l - len(cfg.Data); diff != tc.keysDeleted {
			t.Errorf("expected length of config to be %d, got: %d", l-tc.keysDeleted, len(cfg.Data))
		}
	}
}

func TestConfig_RemoveAll(t *testing.T) {
	data := Data{"/key1": "value1", "/key2": "value2"}
	cfg := CreateConfig(data)
	cfg.RemoveAll()

	if len(cfg.Data) != 0 {
		t.Errorf("expected all keys to be deleted, got %d keys", len(cfg.Data))
	}
}

func TestCreateGlobalConfig(t *testing.T) {
	cfg := CreateConfig(Data{"/key1": "value1"})
	globalCfg := CreateGlobalConfig(cfg)

	if len(globalCfg.Data) != len(cfg.Data) {
		t.Errorf("expected data length %d, got %d", len(cfg.Data), len(globalCfg.Data))
	}
}

func TestCreateDoguConfig(t *testing.T) {
	cfg := CreateConfig(Data{"/key1": "value1"})
	doguCfg := CreateDoguConfig(cfg)

	if len(doguCfg.Data) != len(cfg.Data) {
		t.Errorf("expected data length %d, got %d", len(cfg.Data), len(doguCfg.Data))
	}
}
