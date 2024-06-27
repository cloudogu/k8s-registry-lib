package registry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"testing"
)

type k8sClientTC int

const (
	providerClientWriteConfigMap k8sClientTC = iota
	providerClientWriteConfigMapErr
)

func applyTCForProviderConfigClientMock(tc k8sClientTC, m *MockConfigMapClient) {
	switch tc {
	case providerClientWriteConfigMap:
		m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, ErrConfigNotFound)
		m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(&v1.ConfigMap{}, nil)
	case providerClientWriteConfigMapErr:
		m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, ErrConfigNotFound)
		m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)
	}
}

func applyTCForProviderSecretClientMock(tc k8sClientTC, m *MockSecretClient) {
	switch tc {
	case providerClientWriteConfigMap:
		m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, ErrConfigNotFound)
		m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(&v1.Secret{}, nil)
	case providerClientWriteConfigMapErr:
		m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, ErrConfigNotFound)
		m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)
	}
}

func TestNewDoguConfigRegistryProvider(t *testing.T) {
	t.Run("Provider with ConfigurationRegistry interface", func(t *testing.T) {
		cmClientMock := NewMockConfigMapClient(t)
		applyTCForProviderConfigClientMock(providerClientWriteConfigMap, cmClientMock)

		provider, err := NewDoguConfigRegistryProvider[ConfigurationRegistry](cmClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		_, ok := cfg.(ConfigurationRegistry)
		assert.True(t, ok)
	})

	t.Run("Provider with ConfigurationReader interface", func(t *testing.T) {
		cmClientMock := NewMockConfigMapClient(t)
		applyTCForProviderConfigClientMock(providerClientWriteConfigMap, cmClientMock)

		provider, err := NewDoguConfigRegistryProvider[ConfigurationReader](cmClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		_, ok := cfg.(ConfigurationReader)
		assert.True(t, ok)
	})

	t.Run("Provider with ConfigurationWriter interface", func(t *testing.T) {
		cmClientMock := NewMockConfigMapClient(t)
		applyTCForProviderConfigClientMock(providerClientWriteConfigMap, cmClientMock)

		provider, err := NewDoguConfigRegistryProvider[ConfigurationWriter](cmClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		_, ok := cfg.(ConfigurationWriter)
		assert.True(t, ok)
	})

	t.Run("Provider with custom interface that matches registry", func(t *testing.T) {
		type customRegistry interface {
			Get(ctx context.Context, key string) (string, error)
			Set(ctx context.Context, key, value string) error
		}

		cmClientMock := NewMockConfigMapClient(t)
		applyTCForProviderConfigClientMock(providerClientWriteConfigMap, cmClientMock)

		provider, err := NewDoguConfigRegistryProvider[customRegistry](cmClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		_, ok := cfg.(customRegistry)
		assert.True(t, ok)
	})

	t.Run("Provider with invalid interface that doesn't registry", func(t *testing.T) {
		type invalid interface {
			TEST(ctx context.Context, key string) (int, error)
		}

		cmClientMock := NewMockConfigMapClient(t)

		_, err := NewDoguConfigRegistryProvider[invalid](cmClientMock)
		assert.Error(t, err)
	})

	t.Run("Provider with error while getting config", func(t *testing.T) {
		cmClientMock := NewMockConfigMapClient(t)
		applyTCForProviderConfigClientMock(providerClientWriteConfigMapErr, cmClientMock)

		provider, err := NewDoguConfigRegistryProvider[ConfigurationRegistry](cmClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, cfg)
	})
}

func TestDefaultDoguConfigRegistryProvider_GetDoguConfig(t *testing.T) {
	t.Run("Provider returns default ConfigurationRegistry interface", func(t *testing.T) {
		cmClientMock := NewMockConfigMapClient(t)
		applyTCForProviderConfigClientMock(providerClientWriteConfigMap, cmClientMock)

		provider, err := NewDefaultDoguConfigRegistryProvider(cmClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		_, ok := cfg.(ConfigurationRegistry)
		assert.True(t, ok)
	})

	t.Run("Provider with error while getting config", func(t *testing.T) {
		cmClientMock := NewMockConfigMapClient(t)
		applyTCForProviderConfigClientMock(providerClientWriteConfigMapErr, cmClientMock)

		provider, err := NewDefaultDoguConfigRegistryProvider(cmClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, cfg)
	})
}

func TestNewSensitiveDoguConfigRegistryProvider(t *testing.T) {
	t.Run("Provider with ConfigurationRegistry interface", func(t *testing.T) {
		sClientMock := NewMockSecretClient(t)
		applyTCForProviderSecretClientMock(providerClientWriteConfigMap, sClientMock)

		provider, err := NewSensitiveDoguConfigRegistryProvider[ConfigurationRegistry](sClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		_, ok := cfg.(ConfigurationRegistry)
		assert.True(t, ok)
	})

	t.Run("Provider with ConfigurationReader interface", func(t *testing.T) {
		sClientMock := NewMockSecretClient(t)
		applyTCForProviderSecretClientMock(providerClientWriteConfigMap, sClientMock)

		provider, err := NewSensitiveDoguConfigRegistryProvider[ConfigurationReader](sClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		_, ok := cfg.(ConfigurationReader)
		assert.True(t, ok)
	})

	t.Run("Provider with ConfigurationWriter interface", func(t *testing.T) {
		sClientMock := NewMockSecretClient(t)
		applyTCForProviderSecretClientMock(providerClientWriteConfigMap, sClientMock)

		provider, err := NewSensitiveDoguConfigRegistryProvider[ConfigurationWriter](sClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		_, ok := cfg.(ConfigurationWriter)
		assert.True(t, ok)
	})

	t.Run("Provider with custom interface that matches registry", func(t *testing.T) {
		type customRegistry interface {
			Get(ctx context.Context, key string) (string, error)
			Set(ctx context.Context, key, value string) error
		}

		sClientMock := NewMockSecretClient(t)
		applyTCForProviderSecretClientMock(providerClientWriteConfigMap, sClientMock)

		provider, err := NewSensitiveDoguConfigRegistryProvider[customRegistry](sClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		_, ok := cfg.(customRegistry)
		assert.True(t, ok)
	})

	t.Run("Provider with invalid interface that doesn't registry", func(t *testing.T) {
		type invalid interface {
			TEST(ctx context.Context, key string) (int, error)
		}

		sClientMock := NewMockSecretClient(t)

		_, err := NewSensitiveDoguConfigRegistryProvider[invalid](sClientMock)
		assert.Error(t, err)
	})

	t.Run("Provider with error while getting config", func(t *testing.T) {
		sClientMock := NewMockSecretClient(t)
		applyTCForProviderSecretClientMock(providerClientWriteConfigMapErr, sClientMock)

		provider, err := NewSensitiveDoguConfigRegistryProvider[ConfigurationRegistry](sClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, cfg)
	})
}

func TestDefaultSensitiveDoguConfigRegistryProvider_GetDoguConfig(t *testing.T) {
	t.Run("Provider returns default ConfigurationRegistry interface", func(t *testing.T) {
		sClientMock := NewMockSecretClient(t)
		applyTCForProviderSecretClientMock(providerClientWriteConfigMap, sClientMock)

		provider, err := NewDefaultSensitiveDoguConfigRegistryProvider(sClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		_, ok := cfg.(ConfigurationRegistry)
		assert.True(t, ok)
	})

	t.Run("Provider with error while getting config", func(t *testing.T) {
		sClientMock := NewMockSecretClient(t)
		applyTCForProviderSecretClientMock(providerClientWriteConfigMapErr, sClientMock)

		provider, err := NewDefaultSensitiveDoguConfigRegistryProvider(sClientMock)
		assert.NoError(t, err)

		cfg, err := provider.GetConfig(context.TODO(), "test")
		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, cfg)
	})
}
