package dogu

import (
	_ "embed"
	"encoding/json"
	"github.com/cloudogu/cesapp-lib/core"
	"testing"
)

//go:embed testdata/cas.json
var casBytes []byte

//go:embed testdata/ldap.json
var ldapBytes []byte

func readLdapDogu(t *testing.T) *core.Dogu {
	t.Helper()

	data := &core.Dogu{}
	err := json.Unmarshal(ldapBytes, data)
	if err != nil {
		t.Fatal(err.Error())
	}

	return data
}

func readLdapDoguStr(t *testing.T) string {
	dogu := readLdapDogu(t)
	marshal, err := json.Marshal(dogu)
	if err != nil {
		t.Fatal(err.Error())
	}

	return string(marshal)
}

func readCasDogu(t *testing.T) *core.Dogu {
	t.Helper()

	data := &core.Dogu{}
	err := json.Unmarshal(casBytes, data)
	if err != nil {
		t.Fatal(err.Error())
	}

	return data
}

func readCasDoguStr(t *testing.T) string {
	dogu := readCasDogu(t)
	marshal, err := json.Marshal(dogu)
	if err != nil {
		t.Fatal(err.Error())
	}

	return string(marshal)
}

func parseVersionStr(t *testing.T, version string) core.Version {
	t.Helper()
	parseVersion, err := core.ParseVersion(version)
	if err != nil {
		t.Fatal(err.Error())
	}

	return parseVersion
}
