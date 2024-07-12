package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateDoguConfig(t *testing.T) {
	e := Entries{"key1": "value1"}
	doguName := "test"
	doguCfg := CreateDoguConfig(SimpleDoguName(doguName), e)

	if len(doguCfg.entries) != len(e) {
		t.Errorf("expected data length %d, got %d", len(e), len(doguCfg.entries))
	}

	assert.Equal(t, doguName, doguCfg.DoguName.String())
}
