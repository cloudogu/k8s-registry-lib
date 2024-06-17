package registry

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"testing"
)

type cfgRegistryTC int

const (
	configClientCreateNewConfig cfgRegistryTC = iota
	configClientExistingCfg
	configClientWriteErr
)

func applyTCForConfigClientMock(tc cfgRegistryTC, m *MockConfigMapClient) {
	switch tc {
	case configClientCreateNewConfig:
		m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, ErrConfigNotFound)
		m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(&v1.ConfigMap{}, nil)
	case configClientExistingCfg:
		m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(&v1.ConfigMap{Data: map[string]string{dataKeyName: "key: test"}}, nil)
	case configClientWriteErr:
		m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, ErrConfigNotFound)
		m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("configClientWriteError"))
	}
}

func applyTCForSecretClientMock(tc cfgRegistryTC, m *MockSecretClient) {
	switch tc {
	case configClientCreateNewConfig:
		m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, ErrConfigNotFound)
		m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(&v1.Secret{}, nil)
	case configClientExistingCfg:
		m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(&v1.Secret{Data: map[string][]byte{dataKeyName: []byte("key: test")}}, nil)
	case configClientWriteErr:
		m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, ErrConfigNotFound)
		m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("configClientWriteError"))
	}
}

func TestNewGlobalConfigRegistry(t *testing.T) {
	tests := []struct {
		name string
		tc   cfgRegistryTC
		xErr bool
	}{
		{
			name: "Create initial global config",
			tc:   configClientCreateNewConfig,
			xErr: false,
		},
		{
			name: "Existing initial global config",
			tc:   configClientExistingCfg,
			xErr: false,
		},
		{
			name: "Error writing initial global config",
			tc:   configClientWriteErr,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockConfigMapClient(t)
			applyTCForConfigClientMock(tc.tc, m)

			gcr, err := NewGlobalConfigRegistry(context.TODO(), m)
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				readerRepo := gcr.configReader.repo.(configRepo)
				writerRepo := gcr.configWriter.repo.(configRepo)

				assert.Equal(t, "global", readerRepo.name)
				assert.Equal(t, "global", writerRepo.name)

				assert.Equal(t, m, readerRepo.client.(configMapClient).client)
				assert.Equal(t, m, writerRepo.client.(configMapClient).client)
			}
		})
	}
}

func TestNewDoguConfigRegistry(t *testing.T) {
	tests := []struct {
		name string
		tc   cfgRegistryTC
		xErr bool
	}{
		{
			name: "Create initial dogu config",
			tc:   configClientCreateNewConfig,
			xErr: false,
		},
		{
			name: "Existing initial dogu config",
			tc:   configClientExistingCfg,
			xErr: false,
		},
		{
			name: "Error writing initial dogu config",
			tc:   configClientWriteErr,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockConfigMapClient(t)
			applyTCForConfigClientMock(tc.tc, m)

			gcr, err := NewDoguConfigRegistry(context.TODO(), "myDogu", m)
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				readerRepo := gcr.configReader.repo.(configRepo)
				writerRepo := gcr.configWriter.repo.(configRepo)

				assert.Equal(t, "myDogu-config", readerRepo.name)
				assert.Equal(t, "myDogu-config", writerRepo.name)

				assert.Equal(t, m, readerRepo.client.(configMapClient).client)
				assert.Equal(t, m, writerRepo.client.(configMapClient).client)
			}
		})
	}
}

func TestNewSensitiveDoguRegistry(t *testing.T) {
	tests := []struct {
		name string
		tc   cfgRegistryTC
		xErr bool
	}{
		{
			name: "Create initial sensitive dogu config",
			tc:   configClientCreateNewConfig,
			xErr: false,
		},
		{
			name: "Existing sensitive dogu config",
			tc:   configClientExistingCfg,
			xErr: false,
		},
		{
			name: "Error writing initial sensitive dogu config",
			tc:   configClientWriteErr,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockSecretClient(t)
			applyTCForSecretClientMock(tc.tc, m)

			gcr, err := NewSensitiveDoguRegistry(context.TODO(), "myDogu", m)
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				readerRepo := gcr.configReader.repo.(configRepo)
				writerRepo := gcr.configWriter.repo.(configRepo)

				assert.Equal(t, "myDogu-config", readerRepo.name)
				assert.Equal(t, "myDogu-config", writerRepo.name)

				assert.Equal(t, m, readerRepo.client.(secretClient).client)
				assert.Equal(t, m, writerRepo.client.(secretClient).client)
			}
		})
	}
}

func TestNewGlobalConfigReader(t *testing.T) {
	tests := []struct {
		name string
		tc   cfgRegistryTC
		xErr bool
	}{
		{
			name: "Create initial global config",
			tc:   configClientCreateNewConfig,
			xErr: false,
		},
		{
			name: "Existing initial global config",
			tc:   configClientExistingCfg,
			xErr: false,
		},
		{
			name: "Error writing initial global config",
			tc:   configClientWriteErr,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockConfigMapClient(t)
			applyTCForConfigClientMock(tc.tc, m)

			gcr, err := NewGlobalConfigReader(context.TODO(), m)
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				readerRepo := gcr.configReader.repo.(configRepo)

				assert.Equal(t, "global", readerRepo.name)

				assert.Equal(t, m, readerRepo.client.(configMapClient).client)
			}
		})
	}
}

func TestNewDoguConfigReader(t *testing.T) {
	tests := []struct {
		name string
		tc   cfgRegistryTC
		xErr bool
	}{
		{
			name: "Create initial dogu config",
			tc:   configClientCreateNewConfig,
			xErr: false,
		},
		{
			name: "Existing initial dogu config",
			tc:   configClientExistingCfg,
			xErr: false,
		},
		{
			name: "Error writing initial dogu config",
			tc:   configClientWriteErr,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockConfigMapClient(t)
			applyTCForConfigClientMock(tc.tc, m)

			gcr, err := NewDoguConfigReader(context.TODO(), "myDogu", m)
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				readerRepo := gcr.configReader.repo.(configRepo)

				assert.Equal(t, "myDogu-config", readerRepo.name)

				assert.Equal(t, m, readerRepo.client.(configMapClient).client)
			}
		})
	}
}

func TestNewSensitiveDoguReader(t *testing.T) {
	tests := []struct {
		name string
		tc   cfgRegistryTC
		xErr bool
	}{
		{
			name: "Create initial sensitive dogu config",
			tc:   configClientCreateNewConfig,
			xErr: false,
		},
		{
			name: "Existing initial sensitive dogu config",
			tc:   configClientExistingCfg,
			xErr: false,
		},
		{
			name: "Error writing initial sensitive dogu config",
			tc:   configClientWriteErr,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockSecretClient(t)
			applyTCForSecretClientMock(tc.tc, m)

			gcr, err := NewSensitiveDoguReader(context.TODO(), "myDogu", m)
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				readerRepo := gcr.configReader.repo.(configRepo)

				assert.Equal(t, "myDogu-config", readerRepo.name)

				assert.Equal(t, m, readerRepo.client.(secretClient).client)
			}
		})
	}
}

func TestNewGlobalConfigWatcher(t *testing.T) {
	tests := []struct {
		name string
		tc   cfgRegistryTC
		xErr bool
	}{
		{
			name: "Create initial global config",
			tc:   configClientCreateNewConfig,
			xErr: false,
		},
		{
			name: "Existing initial global config",
			tc:   configClientExistingCfg,
			xErr: false,
		},
		{
			name: "Error writing initial global config",
			tc:   configClientWriteErr,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockConfigMapClient(t)
			applyTCForConfigClientMock(tc.tc, m)

			gcr, err := NewGlobalConfigWatcher(context.TODO(), m)
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				readerRepo := gcr.configWatcher.repo.(configRepo)

				assert.Equal(t, "global", readerRepo.name)

				assert.Equal(t, m, readerRepo.client.(configMapClient).client)
			}
		})
	}
}

func TestNewDoguConfigWatcher(t *testing.T) {
	tests := []struct {
		name string
		tc   cfgRegistryTC
		xErr bool
	}{
		{
			name: "Create initial dogu config",
			tc:   configClientCreateNewConfig,
			xErr: false,
		},
		{
			name: "Existing initial dogu config",
			tc:   configClientExistingCfg,
			xErr: false,
		},
		{
			name: "Error writing initial dogu config",
			tc:   configClientWriteErr,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockConfigMapClient(t)
			applyTCForConfigClientMock(tc.tc, m)

			gcr, err := NewDoguConfigWatcher(context.TODO(), "myDogu", m)
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				readerRepo := gcr.configWatcher.repo.(configRepo)

				assert.Equal(t, "myDogu-config", readerRepo.name)

				assert.Equal(t, m, readerRepo.client.(configMapClient).client)
			}
		})
	}
}

func TestNewSensitiveDoguWatcher(t *testing.T) {
	tests := []struct {
		name string
		tc   cfgRegistryTC
		xErr bool
	}{
		{
			name: "Create initial sensitive dogu config",
			tc:   configClientCreateNewConfig,
			xErr: false,
		},
		{
			name: "Existing initial sensitive dogu config",
			tc:   configClientExistingCfg,
			xErr: false,
		},
		{
			name: "Error writing initial sensitive dogu config",
			tc:   configClientWriteErr,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockSecretClient(t)
			applyTCForSecretClientMock(tc.tc, m)

			gcr, err := NewSensitiveDoguWatcher(context.TODO(), "myDogu", m)
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				readerRepo := gcr.configWatcher.repo.(configRepo)

				assert.Equal(t, "myDogu-config", readerRepo.name)

				assert.Equal(t, m, readerRepo.client.(secretClient).client)
			}
		})
	}
}
