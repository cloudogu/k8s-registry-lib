package registry

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewGlobalConfigRegistry(t *testing.T) {
	t.Run("create global config-registry", func(t *testing.T) {
		mockCmClient := NewMockConfigMapClient(t)

		gcr := NewGlobalConfigRegistry(mockCmClient)

		require.NotNil(t, gcr)

		readerRepo := gcr.configReader.repo.(configRepo)
		writerRepo := gcr.configWriter.repo.(configRepo)

		assert.Equal(t, globalConfigType, readerRepo.configType)
		assert.Equal(t, globalConfigType, writerRepo.configType)

		assert.Equal(t, "global", readerRepo.name)
		assert.Equal(t, "global", writerRepo.name)

		assert.Equal(t, mockCmClient, readerRepo.client.(*configMapClient).client)
		assert.Equal(t, mockCmClient, writerRepo.client.(*configMapClient).client)
	})
}

func TestNewDoguConfigRegistry(t *testing.T) {
	t.Run("create dogu config-registry", func(t *testing.T) {
		mockCmClient := NewMockConfigMapClient(t)

		dcr := NewDoguConfigRegistry("myDogu", mockCmClient)

		require.NotNil(t, dcr)

		readerRepo := dcr.configReader.repo.(configRepo)
		writerRepo := dcr.configWriter.repo.(configRepo)

		assert.Equal(t, doguConfigType, readerRepo.configType)
		assert.Equal(t, doguConfigType, writerRepo.configType)

		assert.Equal(t, "myDogu-config", readerRepo.name)
		assert.Equal(t, "myDogu-config", writerRepo.name)

		assert.Equal(t, mockCmClient, readerRepo.client.(*configMapClient).client)
		assert.Equal(t, mockCmClient, writerRepo.client.(*configMapClient).client)
	})
}

func TestNewSensitiveDoguRegistry(t *testing.T) {
	t.Run("create sensitive config-registry", func(t *testing.T) {
		mockSecretClient := NewMockSecretClient(t)

		sdcr := NewSensitiveDoguRegistry("myDogu", mockSecretClient)

		require.NotNil(t, sdcr)

		readerRepo := sdcr.configReader.repo.(configRepo)
		writerRepo := sdcr.configWriter.repo.(configRepo)

		assert.Equal(t, sensitiveConfigType, readerRepo.configType)
		assert.Equal(t, sensitiveConfigType, writerRepo.configType)

		assert.Equal(t, "myDogu-config", readerRepo.name)
		assert.Equal(t, "myDogu-config", writerRepo.name)

		assert.Equal(t, mockSecretClient, readerRepo.client.(*secretClient).client)
		assert.Equal(t, mockSecretClient, writerRepo.client.(*secretClient).client)
	})
}

func TestNewGlobalConfigReader(t *testing.T) {
	t.Run("create global config-reader", func(t *testing.T) {
		mockCmClient := NewMockConfigMapClient(t)

		gcr := NewGlobalConfigReader(mockCmClient)

		require.NotNil(t, gcr)

		readerRepo := gcr.configReader.repo.(configRepo)
		assert.Equal(t, globalConfigType, readerRepo.configType)
		assert.Equal(t, "global", readerRepo.name)
		assert.Equal(t, mockCmClient, readerRepo.client.(*configMapClient).client)
	})
}

func TestNewDoguConfigReader(t *testing.T) {
	t.Run("create dogu config-reader", func(t *testing.T) {
		mockCmClient := NewMockConfigMapClient(t)

		dcr := NewDoguConfigReader("myDogu", mockCmClient)

		require.NotNil(t, dcr)

		readerRepo := dcr.configReader.repo.(configRepo)
		assert.Equal(t, doguConfigType, readerRepo.configType)
		assert.Equal(t, "myDogu-config", readerRepo.name)
		assert.Equal(t, mockCmClient, readerRepo.client.(*configMapClient).client)
	})
}

func TestNewSensitiveDoguReader(t *testing.T) {
	t.Run("create sensitive config-reader", func(t *testing.T) {
		mockSecretClient := NewMockSecretClient(t)

		sdcr := NewSensitiveDoguReader("myDogu", mockSecretClient)

		require.NotNil(t, sdcr)

		readerRepo := sdcr.configReader.repo.(configRepo)
		assert.Equal(t, sensitiveConfigType, readerRepo.configType)
		assert.Equal(t, "myDogu-config", readerRepo.name)
		assert.Equal(t, mockSecretClient, readerRepo.client.(*secretClient).client)
	})
}
