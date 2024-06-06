package global

import (
	"errors"
	"github.com/stretchr/testify/assert"
	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestConfigSet(t *testing.T) {
	tests := []struct {
		name              string
		etcdSetErr        error
		clusterRegSetErr  error
		expectedError     string
		expectedErrorSubs []string
	}{
		{
			name:             "both successful",
			etcdSetErr:       nil,
			clusterRegSetErr: nil,
			expectedError:    "",
		},
		{
			name:              "only etcd successful",
			etcdSetErr:        nil,
			clusterRegSetErr:  errors.New("testerror"),
			expectedError:     "failed to set key in cluster native registry",
			expectedErrorSubs: []string{"testerror"},
		},
		{
			name:              "only cluster native successful",
			etcdSetErr:        errors.New("testerror"),
			clusterRegSetErr:  nil,
			expectedError:     "failed to set key in etcd registry",
			expectedErrorSubs: []string{"testerror"},
		},
		{
			name:              "both fail",
			etcdSetErr:        errors.New("testerror1"),
			clusterRegSetErr:  errors.New("testerror2"),
			expectedError:     "failed to set key in cluster native registry: testerror2\nfailed to set key in etcd registry: testerror1",
			expectedErrorSubs: []string{"testerror1", "testerror2"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			etcd := &mockEtcdConfigContext{}
			clusterReg := &MockConfigurationRegistry{}
			conf := &Config{
				EtcdRegistry:          etcd,
				ClusterNativeRegistry: clusterReg,
			}

			if tc.etcdSetErr != nil {
				etcd.EXPECT().Set("mykey", "myval").Return(tc.etcdSetErr)
			} else {
				etcd.EXPECT().Set("mykey", "myval").Return(nil)
			}

			if tc.clusterRegSetErr != nil {
				clusterReg.EXPECT().Set(nil, "mykey", "myval").Return(tc.clusterRegSetErr)
			} else {
				clusterReg.EXPECT().Set(nil, "mykey", "myval").Return(nil)
			}

			err := conf.Set(nil, "mykey", "myval")

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)

				for _, sub := range tc.expectedErrorSubs {
					assert.Contains(t, err.Error(), sub)
				}
			} else {
				assert.NoError(t, err) // Check for no error
			}
		})
	}
}

func TestConfigDelete(t *testing.T) {
	tests := []struct {
		name             string
		etcdDeleteErr    error
		clusterRegDelErr error
		expectedError    string
	}{
		{
			name:             "both successful",
			etcdDeleteErr:    nil,
			clusterRegDelErr: nil,
			expectedError:    "",
		},
		{
			name:             "only etcd successful",
			etcdDeleteErr:    nil,
			clusterRegDelErr: errors.New("testerror"),
			expectedError:    "failed to delete key in cluster native registry: testerror",
		},
		{
			name:             "only cluster native successful",
			etcdDeleteErr:    errors.New("testerror"),
			clusterRegDelErr: nil,
			expectedError:    "failed to delete key in etcd registry: testerror",
		},
		{
			name:             "both fail",
			etcdDeleteErr:    errors.New("testerror1"),
			clusterRegDelErr: errors.New("testerror2"),
			expectedError:    "failed to delete key in cluster native registry: testerror2\nfailed to delete key in etcd registry: testerror1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			etcd := &mockEtcdConfigContext{}
			clusterReg := &MockConfigurationRegistry{}
			conf := &Config{
				EtcdRegistry:          etcd,
				ClusterNativeRegistry: clusterReg,
			}

			if tc.etcdDeleteErr != nil {
				etcd.EXPECT().Delete("mykey").Return(tc.etcdDeleteErr)
			} else {
				etcd.EXPECT().Delete("mykey").Return(nil)
			}

			if tc.clusterRegDelErr != nil {
				clusterReg.EXPECT().Delete(nil, "mykey").Return(tc.clusterRegDelErr)
			} else {
				clusterReg.EXPECT().Delete(nil, "mykey").Return(nil)
			}

			err := conf.Delete(nil, "mykey")

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigDeleteRecursive(t *testing.T) {
	tests := []struct {
		name                string
		etcdDeleteRecErr    error
		clusterRegDelRecErr error
		expectedError       string
	}{
		{
			name:                "both successful",
			etcdDeleteRecErr:    nil,
			clusterRegDelRecErr: nil,
			expectedError:       "",
		},
		{
			name:                "only etcd successful",
			etcdDeleteRecErr:    nil,
			clusterRegDelRecErr: errors.New("testerror"),
			expectedError:       "failed to delete recursive in cluster native registry: testerror",
		},
		{
			name:                "only cluster native successful",
			etcdDeleteRecErr:    errors.New("testerror"),
			clusterRegDelRecErr: nil,
			expectedError:       "failed to delete recursive in etcd registry: testerror",
		},
		{
			name:                "both fail",
			etcdDeleteRecErr:    errors.New("testerror1"),
			clusterRegDelRecErr: errors.New("testerror2"),
			expectedError:       "failed to delete recursive in cluster native registry: testerror2\nfailed to delete recursive in etcd registry: testerror1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			etcd := &mockEtcdConfigContext{}
			clusterReg := &MockConfigurationRegistry{}
			conf := &Config{
				EtcdRegistry:          etcd,
				ClusterNativeRegistry: clusterReg,
			}

			if tc.etcdDeleteRecErr != nil {
				etcd.EXPECT().DeleteRecursive("mykey").Return(tc.etcdDeleteRecErr)
			} else {
				etcd.EXPECT().DeleteRecursive("mykey").Return(nil)
			}

			if tc.clusterRegDelRecErr != nil {
				clusterReg.EXPECT().DeleteRecursive(nil, "mykey").Return(tc.clusterRegDelRecErr)
			} else {
				clusterReg.EXPECT().DeleteRecursive(nil, "mykey").Return(nil)
			}

			err := conf.DeleteRecursive(nil, "mykey")

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigRemoveAll(t *testing.T) {
	tests := []struct {
		name                   string
		etcdRemoveAllErr       error
		clusterRegRemoveAllErr error
		expectedError          string
		expectedErrorSubs      []string
	}{
		{
			name:                   "both successful",
			etcdRemoveAllErr:       nil,
			clusterRegRemoveAllErr: nil,
			expectedError:          "",
		},
		{
			name:                   "only etcd successful",
			etcdRemoveAllErr:       nil,
			clusterRegRemoveAllErr: errors.New("testerror"),
			expectedError:          "failed to remove all in cluster native registry",
			expectedErrorSubs:      []string{"testerror"},
		},
		{
			name:                   "only cluster native successful",
			etcdRemoveAllErr:       errors.New("testerror"),
			clusterRegRemoveAllErr: nil,
			expectedError:          "failed to remove all in etcd registry",
			expectedErrorSubs:      []string{"testerror"},
		},
		{
			name:                   "both fail",
			etcdRemoveAllErr:       errors.New("testerror1"),
			clusterRegRemoveAllErr: errors.New("testerror2"),
			expectedError:          "failed to remove all in cluster native registry: testerror2\nfailed to remove all in etcd registry: testerror1",
			expectedErrorSubs:      []string{"testerror1", "testerror2"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			etcd := &mockEtcdConfigContext{}
			clusterReg := &MockConfigurationRegistry{}
			conf := &Config{
				EtcdRegistry:          etcd,
				ClusterNativeRegistry: clusterReg,
			}

			if tc.etcdRemoveAllErr != nil {
				etcd.EXPECT().RemoveAll().Return(tc.etcdRemoveAllErr)
			} else {
				etcd.EXPECT().RemoveAll().Return(nil)
			}

			if tc.clusterRegRemoveAllErr != nil {
				clusterReg.EXPECT().RemoveAll(nil).Return(tc.clusterRegRemoveAllErr)
			} else {
				clusterReg.EXPECT().RemoveAll(nil).Return(nil)
			}

			err := conf.RemoveAll(nil)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)

				for _, sub := range tc.expectedErrorSubs {
					assert.Contains(t, err.Error(), sub)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigGet(t *testing.T) {
	tests := []struct {
		name             string
		etcdGetErr       error
		clusterRegGetErr error
		expectedValue    string
		expectedError    string
	}{
		{
			name:             "value found in clusterReg",
			clusterRegGetErr: nil,
			expectedValue:    "myval",
			expectedError:    "",
		},
		{
			name:             "value found in etcd",
			etcdGetErr:       nil,
			clusterRegGetErr: k8sErrs.NewNotFound(schema.GroupResource{}, ""),
			expectedValue:    "myval",
			expectedError:    "",
		},
		{
			name:             "fail in cluster and etcd",
			etcdGetErr:       errors.New("etcd error"),
			clusterRegGetErr: k8sErrs.NewNotFound(schema.GroupResource{}, ""),
			expectedValue:    "",
			expectedError:    "failed to get key from etcd: etcd error",
		},
		{
			name:             "fail to contact cluster reg",
			clusterRegGetErr: errors.New("critical error"),
			expectedValue:    "",
			expectedError:    "failed to get key from cluster native registry: critical error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			etcd := &mockEtcdConfigContext{}
			clusterReg := &MockConfigurationRegistry{}
			conf := &Config{
				EtcdRegistry:          etcd,
				ClusterNativeRegistry: clusterReg,
			}

			if tc.etcdGetErr != nil {
				etcd.EXPECT().Get("mykey").Return("", tc.etcdGetErr)
			} else {
				etcd.EXPECT().Get("mykey").Return("myval", nil)
			}

			if tc.clusterRegGetErr != nil {
				clusterReg.EXPECT().Get(nil, "mykey").Return("", tc.clusterRegGetErr)
			} else {
				clusterReg.EXPECT().Get(nil, "mykey").Return("myval", nil)
			}

			value, err := conf.Get(nil, "mykey")

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedValue, value)
			}
		})
	}
}

func TestConfigGetAll(t *testing.T) {
	tests := []struct {
		name                string
		etcdGetAllErr       error
		clusterRegGetAllErr error
		expectedValues      map[string]string
		expectedError       string
	}{
		{
			name:                "both successful (values found in clusterReg)",
			etcdGetAllErr:       k8sErrs.NewNotFound(schema.GroupResource{}, ""),
			clusterRegGetAllErr: nil,
			expectedValues:      map[string]string{"key1": "value1", "key2": "value2"},
			expectedError:       "",
		},
		{
			name:                "both successful (values found in etcd)",
			etcdGetAllErr:       nil,
			clusterRegGetAllErr: k8sErrs.NewNotFound(schema.GroupResource{}, ""),
			expectedValues:      map[string]string{"key1": "value1", "key2": "value2"},
			expectedError:       "",
		},
		{
			name:                "fail in cluster and etcd",
			etcdGetAllErr:       errors.New("etcd error"),
			clusterRegGetAllErr: k8sErrs.NewNotFound(schema.GroupResource{}, ""),
			expectedValues:      nil,
			expectedError:       "failed to get all from etcd: etcd error",
		},
		{
			name:                "fail to contact cluster reg",
			clusterRegGetAllErr: errors.New("critical error"),
			expectedValues:      nil,
			expectedError:       "failed to get all from cluster native registry: critical error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			etcd := &mockEtcdConfigContext{}
			clusterReg := &MockConfigurationRegistry{}
			conf := &Config{
				EtcdRegistry:          etcd,
				ClusterNativeRegistry: clusterReg,
			}

			if tc.etcdGetAllErr != nil {
				etcd.EXPECT().GetAll().Return(nil, tc.etcdGetAllErr)
			} else {
				etcd.EXPECT().GetAll().Return(map[string]string{"key1": "value1", "key2": "value2"}, nil)
			}

			if tc.clusterRegGetAllErr != nil {
				clusterReg.EXPECT().GetAll(nil).Return(nil, tc.clusterRegGetAllErr)
			} else {
				clusterReg.EXPECT().GetAll(nil).Return(map[string]string{"key1": "value1", "key2": "value2"}, nil)
			}

			values, err := conf.GetAll(nil)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedValues, values)
			}
		})
	}
}

func TestConfigExists(t *testing.T) {
	tests := []struct {
		name                string
		etcdExistsErr       error
		clusterRegExistsErr error
		expectedExists      bool
		expectedError       string
	}{
		{
			name:                "both successful (key exists in clusterReg)",
			etcdExistsErr:       k8sErrs.NewNotFound(schema.GroupResource{}, ""),
			clusterRegExistsErr: nil,
			expectedExists:      true,
			expectedError:       "",
		},
		{
			name:                "both successful (key exists in etcd)",
			etcdExistsErr:       nil,
			clusterRegExistsErr: k8sErrs.NewNotFound(schema.GroupResource{}, ""),
			expectedExists:      true,
			expectedError:       "",
		},
		{
			name:                "fail in cluster and etcd",
			etcdExistsErr:       errors.New("etcd error"),
			clusterRegExistsErr: k8sErrs.NewNotFound(schema.GroupResource{}, ""),
			expectedExists:      false,
			expectedError:       "failed to read key from etcd: etcd error",
		},
		{
			name:                "fail to contact cluster reg",
			clusterRegExistsErr: errors.New("critical error"),
			expectedExists:      false,
			expectedError:       "failed to read key from cluster native registry: critical error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			etcd := &mockEtcdConfigContext{}
			clusterReg := &MockConfigurationRegistry{}
			conf := &Config{
				EtcdRegistry:          etcd,
				ClusterNativeRegistry: clusterReg,
			}

			if tc.etcdExistsErr != nil {
				etcd.EXPECT().Exists("mykey").Return(false, tc.etcdExistsErr)
			} else {
				etcd.EXPECT().Exists("mykey").Return(true, nil)
			}

			if tc.clusterRegExistsErr != nil {
				clusterReg.EXPECT().Exists(nil, "mykey").Return(false, tc.clusterRegExistsErr)
			} else {
				clusterReg.EXPECT().Exists(nil, "mykey").Return(true, nil)
			}

			exists, err := conf.Exists(nil, "mykey")

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedExists, exists)
			}
		})
	}
}

func TestConfigGetOrFalse(t *testing.T) {
	tests := []struct {
		name                    string
		etcdGetOrFalseErr       error
		clusterRegGetOrFalseErr error
		expectedExists          bool
		expectedValue           string
		expectedError           string
	}{
		{
			name:                    "both successful (key exists in clusterReg)",
			etcdGetOrFalseErr:       k8sErrs.NewNotFound(schema.GroupResource{}, ""),
			clusterRegGetOrFalseErr: nil,
			expectedExists:          true,
			expectedValue:           "myval",
			expectedError:           "",
		},
		{
			name:                    "both successful (key exists in etcd)",
			etcdGetOrFalseErr:       nil,
			clusterRegGetOrFalseErr: k8sErrs.NewNotFound(schema.GroupResource{}, ""),
			expectedExists:          true,
			expectedValue:           "myval",
			expectedError:           "",
		},
		{
			name:                    "fail in cluster and etcd",
			etcdGetOrFalseErr:       errors.New("etcd error"),
			clusterRegGetOrFalseErr: k8sErrs.NewNotFound(schema.GroupResource{}, ""),
			expectedExists:          false,
			expectedValue:           "",
			expectedError:           "failed to get key from etcd: etcd error",
		},
		{
			name:                    "fail to contact cluster reg",
			clusterRegGetOrFalseErr: errors.New("critical error"),
			expectedExists:          false,
			expectedValue:           "",
			expectedError:           "failed to get key from cluster native registry: critical error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			etcd := &mockEtcdConfigContext{}
			clusterReg := &MockConfigurationRegistry{}
			conf := &Config{
				EtcdRegistry:          etcd,
				ClusterNativeRegistry: clusterReg,
			}

			if tc.etcdGetOrFalseErr != nil {
				etcd.EXPECT().GetOrFalse("mykey").Return(false, "", tc.etcdGetOrFalseErr)
			} else {
				etcd.EXPECT().GetOrFalse("mykey").Return(true, "myval", nil)
			}

			if tc.clusterRegGetOrFalseErr != nil {
				clusterReg.EXPECT().GetOrFalse(nil, "mykey").Return(false, "", tc.clusterRegGetOrFalseErr)
			} else {
				clusterReg.EXPECT().GetOrFalse(nil, "mykey").Return(true, "myval", nil)
			}

			exists, value, err := conf.GetOrFalse(nil, "mykey")

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedExists, exists)
				assert.Equal(t, tc.expectedValue, value)
			}
		})
	}
}
