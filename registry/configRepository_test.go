package registry

import (
	"context"
	"errors"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

type configRepo_testcase int

const (
	repo_validReturn configRepo_testcase = iota
	repo_validCreate
	repo_createConverterError
	repo_createClientError
	repo_validUpdate
	repo_updateNoChanges
	repo_updateConverterReadError
	repo_updateConverterWriteError
	repo_updateConfigsEqual
	repo_updateMergeError
	repo_updateClientError
	repo_clientError
	repo_clientGetError
	repo_NotFoundError
	repo_converterError
)

func TestNewConfigRepo(t *testing.T) {
	tests := []struct {
		name     string
		inName   string
		inClient configClient
		xErr     bool
	}{
		{
			name:     "Valid parameters",
			inName:   "test-config",
			inClient: newMockConfigClient(t),
			xErr:     false,
		},
		{
			name:     "empty name",
			inName:   "",
			inClient: newMockConfigClient(t),
			xErr:     true,
		},
		{
			name:     "empty client",
			inName:   "test-config",
			inClient: nil,
			xErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, err := newConfigRepo(tc.inName, tc.inClient)

			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				assert.Equal(t, tc.inName, repo.name)
				assert.Equal(t, tc.inClient, repo.client)
				assert.IsType(t, &config.YamlConverter{}, repo.converter)
			}
		})
	}
}

func TestConfigRepo_get(t *testing.T) {
	applyTestCaseClient := func(m *mockConfigClient, tc configRepo_testcase) {
		switch tc {
		case repo_validReturn, repo_converterError:
			m.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, nil)
		case repo_clientError:
			m.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, errors.New("clientErr"))
		default:
		}
	}

	applyTestCaseConverter := func(m *MockConverter, tc configRepo_testcase) {
		switch tc {
		case repo_validReturn:
			m.EXPECT().Read(mock.Anything).Return(config.Data{}, nil)
		case repo_converterError:
			m.EXPECT().Read(mock.Anything).Return(config.Data{}, errors.New("converterErr"))
		default:
		}
	}

	tests := []struct {
		name   string
		tc     configRepo_testcase
		inName string
		xErr   bool
	}{
		{
			name: "Get",
			tc:   repo_validReturn,
			xErr: false,
		},
		{
			name: "Client Error",
			tc:   repo_clientError,
			xErr: true,
		},
		{
			name: "Converter Error",
			tc:   repo_converterError,
			xErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mClient := newMockConfigClient(t)
			mConverter := NewMockConverter(t)
			applyTestCaseClient(mClient, test.tc)
			applyTestCaseConverter(mConverter, test.tc)

			r := configRepo{
				name:      "testRepo",
				client:    mClient,
				converter: mConverter,
			}

			_, err := r.get(context.TODO())
			assert.Equal(t, test.xErr, err != nil)
		})
	}
}

func TestConfigRepo_delete(t *testing.T) {
	applyTestCaseClient := func(m *mockConfigClient, tc configRepo_testcase) {
		switch tc {
		case repo_validReturn:
			m.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
		case repo_clientError:
			m.EXPECT().Delete(mock.Anything, mock.Anything).Return(errors.New("clientErr"))
		case repo_NotFoundError:
			m.EXPECT().Delete(mock.Anything, mock.Anything).Return(k8serrors.NewNotFound(schema.GroupResource{}, ""))
		default:
		}
	}

	tests := []struct {
		name   string
		tc     configRepo_testcase
		inName string
		xErr   bool
	}{
		{
			name: "Delete",
			tc:   repo_validReturn,
			xErr: false,
		},
		{
			name: "Client Error",
			tc:   repo_clientError,
			xErr: true,
		},
		{
			name: "NotFound Error",
			tc:   repo_NotFoundError,
			xErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mClient := newMockConfigClient(t)
			applyTestCaseClient(mClient, test.tc)

			r := configRepo{
				name:      "testRepo",
				client:    mClient,
				converter: nil,
			}

			err := r.delete(context.TODO())
			assert.Equal(t, test.xErr, err != nil)
		})
	}
}

func TestConfigRepo_write(t *testing.T) {
	applyTestCaseClient := func(m *mockConfigClient, tc configRepo_testcase) {
		switch tc {
		case repo_validCreate:
			m.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, ErrConfigNotFound)
			m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		case repo_createConverterError:
			m.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, ErrConfigNotFound)
		case repo_clientGetError:
			m.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, errors.New("clientErr"))
		case repo_createClientError:
			m.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, ErrConfigNotFound)
			m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("clientErr"))
		case repo_updateNoChanges, repo_updateConverterReadError, repo_updateConfigsEqual, repo_updateConverterWriteError, repo_updateMergeError:
			m.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, nil)
		case repo_updateClientError:
			m.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, nil)
			m.EXPECT().Update(mock.Anything, mock.Anything).Return(errors.New("clientUpdateErr"))
		case repo_validUpdate:
			m.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, nil)
			m.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
		default:
		}
	}

	applyTestCaseConverter := func(m *MockConverter, tc configRepo_testcase) {
		remoteConfig := map[string]string{
			"key1/key2": "keyValue",
		}

		switch tc {
		case repo_validCreate, repo_createClientError:
			m.EXPECT().Write(mock.Anything, mock.Anything).Return(nil)
		case repo_createConverterError:
			m.EXPECT().Write(mock.Anything, mock.Anything).Return(errors.New("converterErr"))
		case repo_updateConverterReadError:
			m.EXPECT().Read(mock.Anything).Return(nil, errors.New("converterErr"))
		case repo_updateConfigsEqual, repo_updateMergeError:
			m.EXPECT().Read(mock.Anything).Return(remoteConfig, nil)
		case repo_updateConverterWriteError:
			m.EXPECT().Read(mock.Anything).Return(remoteConfig, nil)
			m.EXPECT().Write(mock.Anything, mock.Anything).Return(errors.New("converterErr"))
		case repo_updateClientError, repo_validUpdate:
			m.EXPECT().Read(mock.Anything).Return(remoteConfig, nil)
			m.EXPECT().Write(mock.Anything, mock.Anything).Return(nil)
		default:
		}
	}

	tests := []struct {
		name  string
		tc    configRepo_testcase
		inCfg config.Config
		xErr  bool
	}{
		{
			name: "Create new config",
			tc:   repo_validCreate,
			xErr: false,
		},
		{
			name: "Create - converter error",
			tc:   repo_createConverterError,
			xErr: true,
		},
		{
			name: "Create - client error",
			tc:   repo_createClientError,
			xErr: true,
		},
		{
			name: "Update",
			tc:   repo_validUpdate,
			inCfg: config.Config{
				Data: map[string]string{
					"key1/key2": "newKeyValue",
				},
				ChangeHistory: []config.Change{
					{
						KeyPath: "key1/key2",
						Deleted: false,
					},
				},
			},
			xErr: false,
		},
		{
			name: "Update - no changes",
			tc:   repo_updateNoChanges,
			inCfg: config.Config{
				Data:          make(config.Data),
				ChangeHistory: make([]config.Change, 0),
			},
			xErr: false,
		},
		{
			name: "Update - converter read error",
			tc:   repo_updateConverterReadError,
			inCfg: config.Config{
				Data: map[string]string{
					"key1/key2": "keyValue",
				},
				ChangeHistory: []config.Change{
					{
						KeyPath: "key1/key2",
						Deleted: false,
					},
				},
			},
			xErr: true,
		},
		{
			name: "Update - equal configs",
			tc:   repo_updateConfigsEqual,
			inCfg: config.Config{
				Data: map[string]string{
					"key1/key2": "keyValue",
				},
				ChangeHistory: []config.Change{
					{
						KeyPath: "key1/key2",
						Deleted: false,
					},
				},
			},
			xErr: false,
		},
		{
			name: "Update - merge error",
			tc:   repo_updateMergeError,
			inCfg: config.Config{
				Data: map[string]string{
					"key11":     "keyValue11",
					"key1/key2": "keyValue",
				},
				ChangeHistory: []config.Change{
					{
						KeyPath: "key1",
						Deleted: false,
					},
				},
			},
			xErr: true,
		},
		{
			name: "Update - converter write error after merge",
			tc:   repo_updateConverterWriteError,
			inCfg: config.Config{
				Data: map[string]string{
					"key1/key2": "newKeyValue",
				},
				ChangeHistory: []config.Change{
					{
						KeyPath: "key1/key2",
						Deleted: false,
					},
				},
			},
			xErr: true,
		},
		{
			name: "Update - client update error after merge",
			tc:   repo_updateClientError,
			inCfg: config.Config{
				Data: map[string]string{
					"key1/key2": "newKeyValue",
				},
				ChangeHistory: []config.Change{
					{
						KeyPath: "key1/key2",
						Deleted: false,
					},
				},
			},
			xErr: true,
		},
		{
			name: "Error getting remote config",
			tc:   repo_clientGetError,
			xErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mClient := newMockConfigClient(t)
			mConverter := NewMockConverter(t)
			applyTestCaseClient(mClient, test.tc)
			applyTestCaseConverter(mConverter, test.tc)

			r := configRepo{
				name:      "testRepo",
				client:    mClient,
				converter: mConverter,
			}

			err := r.write(context.TODO(), test.inCfg)
			assert.Equal(t, test.xErr, err != nil)
		})
	}
}

func TestMergeConfigData(t *testing.T) {
	tests := []struct {
		name      string
		remoteCfg config.Data
		localCfg  config.Config
		xErr      bool
		xResult   config.Data
	}{
		{
			name: "local config - key added",
			remoteCfg: map[string]string{
				"key1": "value1",
			},
			localCfg: config.Config{
				Data: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				ChangeHistory: []config.Change{
					{
						KeyPath: "key2",
						Deleted: false,
					},
				},
			},
			xErr: false,
			xResult: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "local config - key deleted",
			remoteCfg: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			localCfg: config.Config{
				Data: map[string]string{
					"key1": "value1",
				},
				ChangeHistory: []config.Change{
					{
						KeyPath: "key2",
						Deleted: true,
					},
				},
			},
			xErr: false,
			xResult: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "local config - key overridden",
			remoteCfg: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			localCfg: config.Config{
				Data: map[string]string{
					"key1": "value1",
					"key2": "newValue",
				},
				ChangeHistory: []config.Change{
					{
						KeyPath: "key2",
						Deleted: false,
					},
				},
			},
			xErr: false,
			xResult: map[string]string{
				"key1": "value1",
				"key2": "newValue",
			},
		},
		{
			name: "remote config - key added",
			remoteCfg: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			localCfg: config.Config{
				Data: map[string]string{
					"key1": "newValue",
				},
				ChangeHistory: []config.Change{
					{
						KeyPath: "key1",
						Deleted: false,
					},
				},
			},
			xErr: false,
			xResult: map[string]string{
				"key1": "newValue",
				"key2": "value2",
			},
		},
		{
			name: "remote config - key deleted",
			remoteCfg: map[string]string{
				"key1": "value1",
				"key3": "value3",
			},
			localCfg: config.Config{
				Data: map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "newValue",
				},
				ChangeHistory: []config.Change{
					{
						KeyPath: "key3",
						Deleted: false,
					},
				},
			},
			xErr: false,
			xResult: map[string]string{
				"key1": "value1",
				"key3": "newValue",
			},
		},
		{
			name: "remote config - merge conflict - remote key2 delete - local key2 changed",
			remoteCfg: map[string]string{
				"key1": "value1",
				"key3": "remoteNewValue",
			},
			localCfg: config.Config{
				Data: map[string]string{
					"key1": "value1",
					"key2": "newValue2",
					"key3": "newValue3",
				},
				ChangeHistory: []config.Change{
					{
						KeyPath: "key3",
						Deleted: false,
					},
					{
						KeyPath: "key2",
						Deleted: false,
					},
				},
			},
			xErr: false,
			xResult: map[string]string{
				"key1": "value1",
				"key2": "newValue2",
				"key3": "newValue3",
			},
		},
		{
			name: "local config get error",
			remoteCfg: map[string]string{
				"key1": "value1",
			},
			localCfg: config.Config{
				Data: map[string]string{
					"key1": "newValue",
				},
				ChangeHistory: []config.Change{
					{
						KeyPath: "key3",
						Deleted: false,
					},
				},
			},
			xErr:    true,
			xResult: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := mergeConfigData(tc.remoteCfg, tc.localCfg)
			assert.Equal(t, tc.xErr, err != nil)
			assert.Equal(t, tc.xResult, res)
		})
	}
}
