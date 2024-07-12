package config

import "testing"

func TestCreateGlobalConfig(t *testing.T) {
	e := Entries{"key1": "value1"}
	globalCfg := CreateGlobalConfig(e)

	if len(globalCfg.entries) != len(e) {
		t.Errorf("expected data length %d, got %d", len(e), len(globalCfg.entries))
	}
}
