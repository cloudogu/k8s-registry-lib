package repository

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

type configRepo_testcase int

const (
	repo_validReturn configRepo_testcase = iota
	repo_validUpdate
	repo_updateNoChanges
	repo_updateConverterReadError
	repo_updateConverterWriteError
	repo_updateConfigsEqual
	repo_updateClientError
	repo_clientError
	repo_clientGetError
	repo_converterError
)

func TestNewConfigRepo(t *testing.T) {
	tests := []struct {
		name     string
		inClient configClient
		xErr     bool
	}{
		{
			name:     "Valid parameters",
			inClient: newMockConfigClient(t),
			xErr:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newConfigRepo(tc.inClient)

			assert.Equal(t, tc.inClient, repo.client)
			assert.IsType(t, &config.YamlConverter{}, repo.converter)
		})
	}
}

func TestConfigRepo_get(t *testing.T) {
	applyTestCaseClient := func(m *mockConfigClient, tc configRepo_testcase) {
		switch tc {
		case repo_validReturn, repo_converterError:
			m.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, nil)
		case repo_clientError:
			m.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, assert.AnError)
		default:
		}
	}

	applyTestCaseConverter := func(m *mockConverter, tc configRepo_testcase) {
		switch tc {
		case repo_validReturn:
			m.EXPECT().Read(mock.Anything).Return(config.Entries{}, nil)
		case repo_converterError:
			m.EXPECT().Read(mock.Anything).Return(config.Entries{}, assert.AnError)
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
			mConverter := newMockConverter(t)
			applyTestCaseClient(mClient, test.tc)
			applyTestCaseConverter(mConverter, test.tc)

			r := configRepository{
				client:    mClient,
				converter: mConverter,
			}

			_, err := r.get(context.TODO(), createConfigName("test"))
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
			m.EXPECT().Delete(mock.Anything, mock.Anything).Return(assert.AnError)
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mClient := newMockConfigClient(t)
			applyTestCaseClient(mClient, test.tc)

			r := configRepository{
				client:    mClient,
				converter: nil,
			}

			err := r.delete(context.TODO(), createConfigName("test"))
			assert.Equal(t, test.xErr, err != nil)
		})
	}
}

func TestConfigRepo_create(t *testing.T) {
	applyTestCases := func(mClient *mockConfigClient, mConverter *mockConverter, tc configRepo_testcase) {
		switch tc {
		case repo_validReturn:
			mConverter.EXPECT().Write(mock.Anything, mock.Anything).Return(nil)

			mGetter := newMockResourceVersionGetter(t)
			mGetter.EXPECT().GetResourceVersion().Return(resourceVersion)
			mClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mGetter, nil)
		case repo_clientError:
			mConverter.EXPECT().Write(mock.Anything, mock.Anything).Return(nil)

			mClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)
		case repo_converterError:
			mConverter.EXPECT().Write(mock.Anything, mock.Anything).Return(assert.AnError)
		default:
		}
	}

	tests := []struct {
		name    string
		tc      configRepo_testcase
		xErr    bool
		xResult string
	}{
		{
			name:    "Create config",
			tc:      repo_validReturn,
			xErr:    false,
			xResult: resourceVersion,
		},
		{
			name: "Converter Error",
			tc:   repo_converterError,
			xErr: true,
		},
		{
			name: "Client Error",
			tc:   repo_clientError,
			xErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mClient := newMockConfigClient(t)
			mConverter := newMockConverter(t)

			applyTestCases(mClient, mConverter, tt.tc)

			r := configRepository{
				client:    mClient,
				converter: mConverter,
			}

			res, err := r.create(context.TODO(), "", "", config.Config{})
			assert.Equal(t, tt.xErr, err != nil)

			if err == nil {
				assert.Equal(t, tt.xResult, res.PersistenceContext)
			}
		})
	}
}

func TestConfigRepo_update(t *testing.T) {
	applyTestCases := func(mClient *mockConfigClient, mConverter *mockConverter, tc configRepo_testcase) {
		switch tc {
		case repo_validReturn:
			mConverter.EXPECT().Write(mock.Anything, mock.Anything).Return(nil)

			mGetter := newMockResourceVersionGetter(t)
			mGetter.EXPECT().GetResourceVersion().Return(resourceVersion)
			mClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mGetter, nil)
		case repo_clientError:
			mConverter.EXPECT().Write(mock.Anything, mock.Anything).Return(nil)

			mClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)
		case repo_converterError:
			mConverter.EXPECT().Write(mock.Anything, mock.Anything).Return(assert.AnError)
		default:
		}
	}

	tests := []struct {
		name    string
		tc      configRepo_testcase
		xErr    bool
		xResult string
	}{
		{
			name:    "Create config",
			tc:      repo_validReturn,
			xErr:    false,
			xResult: resourceVersion,
		},
		{
			name: "Converter Error",
			tc:   repo_converterError,
			xErr: true,
		},
		{
			name: "Client Error",
			tc:   repo_clientError,
			xErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mClient := newMockConfigClient(t)
			mConverter := newMockConverter(t)

			applyTestCases(mClient, mConverter, tt.tc)

			r := configRepository{
				client:    mClient,
				converter: mConverter,
			}

			res, err := r.update(context.TODO(), "", "", config.Config{})
			assert.Equal(t, tt.xErr, err != nil)

			if err == nil {
				assert.Equal(t, tt.xResult, res.PersistenceContext)
			}
		})
	}
}

func TestConfigRepo_write(t *testing.T) {
	applyTestCaseClient := func(mClient *mockConfigClient, mConverter *mockConverter, tc configRepo_testcase) {
		remoteConfig := map[config.Key]config.Value{
			"key1/key2": "keyValue",
		}

		switch tc {
		case repo_validUpdate:
			mGetter := newMockResourceVersionGetter(t)
			mGetter.EXPECT().GetResourceVersion().Return(resourceVersion)

			mClient.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, nil)
			mClient.EXPECT().UpdateClientData(mock.Anything, mock.Anything).Return(mGetter, nil)

			mConverter.EXPECT().Read(mock.Anything).Return(remoteConfig, nil)
			mConverter.EXPECT().Write(mock.Anything, mock.Anything).Return(nil)
		case repo_clientGetError:
			mClient.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, assert.AnError)
		case repo_updateConverterReadError:
			mClient.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, nil)
			mConverter.EXPECT().Read(mock.Anything).Return(nil, assert.AnError)
		case repo_updateConfigsEqual:
			mClient.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, nil)
			mConverter.EXPECT().Read(mock.Anything).Return(remoteConfig, nil)
		case repo_updateConverterWriteError:
			mClient.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, nil)
			mConverter.EXPECT().Read(mock.Anything).Return(remoteConfig, nil)
			mConverter.EXPECT().Write(mock.Anything, mock.Anything).Return(assert.AnError)
		case repo_updateClientError:
			mClient.EXPECT().Get(mock.Anything, mock.Anything).Return(clientData{}, nil)
			mConverter.EXPECT().Read(mock.Anything).Return(remoteConfig, nil)
			mConverter.EXPECT().Write(mock.Anything, mock.Anything).Return(nil)
			mClient.EXPECT().UpdateClientData(mock.Anything, mock.Anything).Return(nil, assert.AnError)
		default:
		}
	}

	tests := []struct {
		name    string
		tc      configRepo_testcase
		inCfg   config.Config
		xErr    bool
		xResult string
	}{
		{
			name: "UpdateClientData",
			tc:   repo_validUpdate,
			inCfg: createConfigWithChanges(t,
				map[config.Key]config.Value{
					"key1/key2": "newKeyValue",
				},
				[]config.Change{
					{
						KeyPath: "key1/key2",
						Deleted: false,
					},
				}),
			xErr:    false,
			xResult: resourceVersion,
		},
		{
			name: "UpdateClientData - no changes",
			tc:   repo_updateNoChanges,
			inCfg: createConfigWithChanges(t,
				make(config.Entries),
				make([]config.Change, 0)),
			xErr:    false,
			xResult: "",
		},
		{
			name: "Client Get Error",
			tc:   repo_clientGetError,
			inCfg: createConfigWithChanges(t,
				map[config.Key]config.Value{
					"key1/key2": "newKeyValue",
				},
				[]config.Change{
					{
						KeyPath: "key1/key2",
						Deleted: false,
					},
				}),
			xErr: true,
		},
		{
			name: "UpdateClientData - converter read error",
			tc:   repo_updateConverterReadError,
			inCfg: createConfigWithChanges(t,
				map[config.Key]config.Value{
					"key1/key2": "keyValue",
				},
				[]config.Change{
					{
						KeyPath: "key1/key2",
						Deleted: false,
					},
				}),
			xErr: true,
		},
		{
			name: "UpdateClientData - equal configs",
			tc:   repo_updateConfigsEqual,
			inCfg: createConfigWithChanges(t,
				map[config.Key]config.Value{
					"key1/key2": "keyValue",
				},
				[]config.Change{
					{
						KeyPath: "key1/key2",
						Deleted: false,
					},
				}),
			xErr: false,
		},
		{
			name: "UpdateClientData - converter write error after merge",
			tc:   repo_updateConverterWriteError,
			inCfg: createConfigWithChanges(t,
				map[config.Key]config.Value{
					"key1/key2": "newKeyValue",
				},
				[]config.Change{
					{
						KeyPath: "key1/key2",
						Deleted: false,
					},
				}),
			xErr: true,
		},
		{
			name: "UpdateClientData - client update error after merge",
			tc:   repo_updateClientError,
			inCfg: createConfigWithChanges(t,
				map[config.Key]config.Value{
					"key1/key2": "newKeyValue",
				},
				[]config.Change{
					{
						KeyPath: "key1/key2",
						Deleted: false,
					},
				}),
			xErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mClient := newMockConfigClient(t)
			mConverter := newMockConverter(t)
			applyTestCaseClient(mClient, mConverter, test.tc)

			r := configRepository{
				client:    mClient,
				converter: mConverter,
			}

			uCfg, err := r.saveOrMerge(context.TODO(), "", test.inCfg)
			assert.Equal(t, test.xErr, err != nil)

			if err == nil {
				assert.Equal(t, test.xResult, uCfg.PersistenceContext)
			}
		})
	}
}

func TestMergeConfigData(t *testing.T) {
	tests := []struct {
		name      string
		remoteCfg config.Entries
		localCfg  config.Config
		xErr      bool
		xResult   config.Entries
	}{
		{
			name: "local config - key added",
			remoteCfg: map[config.Key]config.Value{
				"key1": "value1",
			},
			localCfg: createConfigWithChanges(t,
				map[config.Key]config.Value{
					"key1": "value1",
					"key2": "value2",
				},
				[]config.Change{
					{
						KeyPath: "key2",
						Deleted: false,
					},
				}),
			xErr: false,
			xResult: map[config.Key]config.Value{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "local config - key deleted",
			remoteCfg: map[config.Key]config.Value{
				"key1": "value1",
				"key2": "value2",
			},
			localCfg: createConfigWithChanges(t,
				map[config.Key]config.Value{
					"key1": "value1",
					"key2": "value2",
				},
				[]config.Change{
					{
						KeyPath: "key2",
						Deleted: true,
					},
				}),
			xErr: false,
			xResult: map[config.Key]config.Value{
				"key1": "value1",
			},
		},
		{
			name: "local config - key overridden",
			remoteCfg: map[config.Key]config.Value{
				"key1": "value1",
				"key2": "value2",
			},
			localCfg: createConfigWithChanges(t,
				map[config.Key]config.Value{
					"key1": "value1",
					"key2": "newValue",
				},
				[]config.Change{
					{
						KeyPath: "key2",
						Deleted: false,
					},
				}),
			xErr: false,
			xResult: map[config.Key]config.Value{
				"key1": "value1",
				"key2": "newValue",
			},
		},
		{
			name: "remote config - key added",
			remoteCfg: map[config.Key]config.Value{
				"key1": "value1",
				"key2": "value2",
			},
			localCfg: createConfigWithChanges(t,
				map[config.Key]config.Value{
					"key1": "newValue",
				},
				[]config.Change{
					{
						KeyPath: "key1",
						Deleted: false,
					},
				}),
			xErr: false,
			xResult: map[config.Key]config.Value{
				"key1": "newValue",
				"key2": "value2",
			},
		},
		{
			name: "remote config - key deleted",
			remoteCfg: map[config.Key]config.Value{
				"key1": "value1",
				"key3": "value3",
			},
			localCfg: createConfigWithChanges(t,
				map[config.Key]config.Value{
					"key1": "value1",
					"key2": "value2",
					"key3": "newValue",
				},
				[]config.Change{
					{
						KeyPath: "key3",
						Deleted: false,
					},
				}),
			xErr: false,
			xResult: map[config.Key]config.Value{
				"key1": "value1",
				"key3": "newValue",
			},
		},
		{
			name: "remote config - merge conflict - remote key2 delete - local key2 changed",
			remoteCfg: map[config.Key]config.Value{
				"key1": "value1",
				"key3": "remoteNewValue",
			},
			localCfg: createConfigWithChanges(t,
				map[config.Key]config.Value{
					"key1": "value1",
					"key2": "newValue2",
					"key3": "newValue3",
				},
				[]config.Change{
					{
						KeyPath: "key3",
						Deleted: false,
					},
					{
						KeyPath: "key2",
						Deleted: false,
					},
				}),
			xErr: false,
			xResult: map[config.Key]config.Value{
				"key1": "value1",
				"key2": "newValue2",
				"key3": "newValue3",
			},
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

func Test_configRepo_watch(t *testing.T) {
	ctx := context.TODO()

	t.Run("should watch config for any changes", func(t *testing.T) {
		resultChan := make(chan clientWatchResult)

		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		mockClient := newMockConfigClient(t)
		mockClient.EXPECT().Watch(ctxTimeout, "dogu-config").Return(resultChan, nil)

		repo := newConfigRepo(mockClient)

		var wg sync.WaitGroup

		wgDone := make(chan struct{})
		defer close(wgDone)

		watch, err := repo.watch(ctxTimeout, "dogu-config")
		assert.NoError(t, err)

		wg.Add(1)
		go func() {
			defer wg.Done()

			resultChan <- clientWatchResult{configRaw{"foo: bar", ""}, configRaw{"foo: value", ""}, nil}
			resultChan <- clientWatchResult{configRaw{"foo: value", ""}, configRaw{"key: other", ""}, nil}
			resultChan <- clientWatchResult{configRaw{}, configRaw{}, assert.AnError}

			close(resultChan)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			i := 0
			for result := range watch {
				if i == 0 {
					assert.NoError(t, result.err)
					assert.Equal(t, config.CreateConfig(map[config.Key]config.Value{"foo": "bar"}, config.WithPersistenceContext("")), result.prevState)
					assert.Equal(t, config.CreateConfig(map[config.Key]config.Value{"foo": "value"}, config.WithPersistenceContext("")), result.newState)
				}

				if i == 1 {
					assert.NoError(t, result.err)
					assert.Equal(t, config.CreateConfig(map[config.Key]config.Value{"foo": "value"}, config.WithPersistenceContext("")), result.prevState)
					assert.Equal(t, config.CreateConfig(map[config.Key]config.Value{"key": "other"}, config.WithPersistenceContext("")), result.newState)
				}

				if i == 2 {
					assert.ErrorIs(t, result.err, assert.AnError)
					assert.ErrorContains(t, result.err, "client watch error:")
				}

				i++
			}
		}()

		go func() {
			wg.Wait()
			wgDone <- struct{}{}
		}()

		select {
		case <-wgDone:
		case <-ctxTimeout.Done():
			t.Errorf("did not reach all evente in time")
		}
	})

	t.Run("should watch config for changes that matches filter", func(t *testing.T) {
		resultChan := make(chan clientWatchResult)

		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		mockClient := newMockConfigClient(t)
		mockClient.EXPECT().Watch(ctxTimeout, "dogu-config").Return(resultChan, nil)

		repo := newConfigRepo(mockClient)

		var wg sync.WaitGroup

		wgDone := make(chan struct{})
		defer close(wgDone)

		watch, err := repo.watch(ctxTimeout, "dogu-config", config.KeyFilter("key"))
		assert.NoError(t, err)

		wg.Add(1)
		go func() {
			defer wg.Done()

			resultChan <- clientWatchResult{configRaw{"foo: bar", ""}, configRaw{"foo: value", ""}, nil}
			resultChan <- clientWatchResult{configRaw{"foo: value", ""}, configRaw{"key: other", ""}, nil}
			resultChan <- clientWatchResult{configRaw{}, configRaw{}, assert.AnError}

			close(resultChan)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			i := 0
			for result := range watch {
				if i == 0 {
					assert.NoError(t, result.err)
					assert.Equal(t, config.CreateConfig(map[config.Key]config.Value{"foo": "value"}, config.WithPersistenceContext("")), result.prevState)
					assert.Equal(t, config.CreateConfig(map[config.Key]config.Value{"key": "other"}, config.WithPersistenceContext("")), result.newState)
				}

				if i == 1 {
					assert.ErrorIs(t, result.err, assert.AnError)
					assert.ErrorContains(t, result.err, "client watch error:")
				}

				i++
			}
		}()

		go func() {
			wg.Wait()
			wgDone <- struct{}{}
		}()

		select {
		case <-wgDone:
		case <-ctxTimeout.Done():
			t.Errorf("did not reach all evente in time")
		}
	})

	t.Run("should not notify when filter is not matching", func(t *testing.T) {
		resultChan := make(chan clientWatchResult)

		ctxTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		mockClient := newMockConfigClient(t)
		mockClient.EXPECT().Watch(ctxTimeout, "dogu-config").Return(resultChan, nil)

		repo := newConfigRepo(mockClient)

		var wg sync.WaitGroup

		wgDone := make(chan struct{})
		defer close(wgDone)

		watch, err := repo.watch(ctxTimeout, "dogu-config", config.KeyFilter("n/a"))
		assert.NoError(t, err)

		wg.Add(1)
		go func() {
			defer wg.Done()

			resultChan <- clientWatchResult{configRaw{"foo: bar", ""}, configRaw{"foo: value", ""}, nil}
			resultChan <- clientWatchResult{configRaw{"foo: value", ""}, configRaw{"key: other", ""}, nil}

			close(resultChan)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			for range watch {
				assert.Fail(t, "unexpected result received, expected no notification")
			}
		}()

		go func() {
			wg.Wait()
			wgDone <- struct{}{}
		}()

		select {
		case <-wgDone:
		case <-ctxTimeout.Done():
			t.Errorf("did not reach all evente in time")
		}
	})
}

func Test_createConfigWatchResult(t *testing.T) {
	t.Run("should fail to create configWatchResult for error reading config", func(t *testing.T) {
		watchResult := clientWatchResult{
			prevConfig: configRaw{
				data: "noValidYaml!!",
			},
		}

		result := createConfigWatchResult(watchResult, &config.YamlConverter{})

		assert.Equal(t, config.Config{}, result.prevState)
		assert.Equal(t, config.Config{}, result.newState)
		assert.Error(t, result.err)
		assert.ErrorContains(t, result.err, "could not convert previous state to config: could not convert client data to config data: unable to decode yaml from reader:")

	})
}

func createConfigWithChanges(t *testing.T, initialEntries config.Entries, changes []config.Change) config.Config {
	cfg := config.CreateConfig(initialEntries, config.WithPersistenceContext(""))

	for _, c := range changes {
		if c.Deleted {
			cfg = cfg.Delete(c.KeyPath)
			continue
		}

		sCfg, err := cfg.Set(c.KeyPath, initialEntries[c.KeyPath])
		require.NoError(t, err)

		cfg = sCfg
	}

	return cfg
}
