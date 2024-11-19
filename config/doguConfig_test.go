package config

import (
	"github.com/cloudogu/ces-commons-lib/dogu"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateDoguConfig(t *testing.T) {
	e := Entries{"key1": "value1"}
	doguName := "test"
	doguCfg := CreateDoguConfig(dogu.SimpleName(doguName), e)

	if len(doguCfg.entries) != len(e) {
		t.Errorf("expected data length %d, got %d", len(e), len(doguCfg.entries))
	}

	assert.Equal(t, doguName, doguCfg.DoguName.String())
}
