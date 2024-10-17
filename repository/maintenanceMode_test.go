package repository

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/cloudogu/k8s-registry-lib/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

var testCtx = context.Background()

func Test_defaultSwitcher_Activate(t *testing.T) {
	t.Run("should fail to activate maintenance mode", func(t *testing.T) {
		// given
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(config.Config{}, assert.AnError)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		sut := &MaintenanceModeAdapter{
			owner:            "k8s-blueprint-operator",
			globalConfigRepo: repo,
		}

		// when
		err := sut.Activate(testCtx, MaintenanceModeDescription{
			Title: "myTitle",
			Text:  "myText",
		})

		// then
		require.Error(t, err)
		assert.True(t, errors.IsGenericError(err))
		assert.ErrorContains(t, err, "could not get contents of global config-map for activating maintenance mode")
	})

	t.Run("should activate maintenance mode if it is currently not active", func(t *testing.T) {
		// given
		expectedJson := `{"title":"myTitle","text":"myText","holder":"k8s-blueprint-operator"}`
		mConfigRepo := newMockGeneralConfigRepository(t)
		globalConfig := config.CreateConfig(config.Entries{})
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(globalConfig, nil)
		mConfigRepo.EXPECT().update(testCtx, configName("global-config"), config.SimpleDoguName(""), mock.Anything).Return(globalConfig, nil)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		sut := &MaintenanceModeAdapter{
			owner:            "k8s-blueprint-operator",
			globalConfigRepo: repo,
		}

		// when
		err := sut.Activate(testCtx, MaintenanceModeDescription{
			Title: "myTitle",
			Text:  "myText",
		})

		// then
		require.NoError(t, err)
		title, b := globalConfig.Get("maintenance")
		require.True(t, b)
		assert.Equal(t, config.Value(expectedJson), title)
	})

	t.Run("should return error on error updating global config", func(t *testing.T) {
		// given
		mConfigRepo := newMockGeneralConfigRepository(t)
		globalConfig := config.CreateConfig(config.Entries{})
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(globalConfig, nil)
		mConfigRepo.EXPECT().update(testCtx, configName("global-config"), config.SimpleDoguName(""), mock.Anything).Return(config.Config{}, assert.AnError)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		sut := &MaintenanceModeAdapter{
			owner:            "k8s-blueprint-operator",
			globalConfigRepo: repo,
		}

		// when
		err := sut.Activate(testCtx, MaintenanceModeDescription{
			Title: "myTitle",
			Text:  "myText",
		})

		// then
		require.Error(t, err)
		require.ErrorContains(t, err, "could not update global config-map for activating maintenance mode")
		assert.True(t, errors.IsGenericError(err))
	})

	t.Run("should return conflict error if the maintenance is activated from another user", func(t *testing.T) {
		// given
		mConfigRepo := newMockGeneralConfigRepository(t)
		globalConfig := config.CreateConfig(config.Entries{"maintenance": "{\"title\": \"title\", \"text\": \"text\", \"holder\": \"k8s-backup-operator\"}"})
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(globalConfig, nil)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		sut := &MaintenanceModeAdapter{
			owner:            "k8s-blueprint-operator",
			globalConfigRepo: repo,
		}

		// when
		err := sut.Activate(testCtx, MaintenanceModeDescription{
			Title: "myTitle",
			Text:  "myText",
		})

		// then
		require.Error(t, err)
		assert.ErrorContains(t, err, "maintenance mode {\"title\": \"title\", \"text\": \"text\", \"holder\": \"k8s-backup-operator\"} is already activated by another owner: k8s-backup-operator")
		assert.True(t, errors.IsConflictError(err))
	})

	t.Run("should return generic error if the actual maintenance mode value can't be parsed", func(t *testing.T) {
		// given
		mConfigRepo := newMockGeneralConfigRepository(t)
		globalConfig := config.CreateConfig(config.Entries{"maintenance": "{\"title\": 1}"})
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(globalConfig, nil)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		sut := &MaintenanceModeAdapter{
			owner:            "k8s-blueprint-operator",
			globalConfigRepo: repo,
		}

		// when
		err := sut.Activate(testCtx, MaintenanceModeDescription{
			Title: "myTitle",
			Text:  "myText",
		})

		// then
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to parse json of maintenance mode config")
		assert.True(t, errors.IsGenericError(err))
	})

	t.Run("should return nil and do nothing if a component already holds the maintenance mode", func(t *testing.T) {
		// given
		mConfigRepo := newMockGeneralConfigRepository(t)
		globalConfig := config.CreateConfig(config.Entries{"maintenance": "{\"title\": \"title\", \"text\": \"text\", \"holder\": \"k8s-blueprint-operator\"}"})
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(globalConfig, nil)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		sut := &MaintenanceModeAdapter{
			owner:            "k8s-blueprint-operator",
			globalConfigRepo: repo,
		}

		// when
		err := sut.Activate(testCtx, MaintenanceModeDescription{
			Title: "myTitle",
			Text:  "myText",
		})

		// then
		require.NoError(t, err)
	})
}

func TestSwitch_Deactivate(t *testing.T) {
	t.Run("should fail to deactivate maintenance mode", func(t *testing.T) {
		// given
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(config.Config{}, assert.AnError)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		sut := &MaintenanceModeAdapter{
			owner:            "k8s-blueprint-operator",
			globalConfigRepo: repo,
		}

		// when
		err := sut.Deactivate(testCtx)

		// then
		require.Error(t, err)
		assert.True(t, errors.IsGenericError(err))
		assert.ErrorContains(t, err, "could not get contents of global config-map for deactivating maintenance mode")
	})
	t.Run("should do nothing if the maintenance mode is not activated", func(t *testing.T) {
		// given
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(config.Config{}, nil)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		sut := &MaintenanceModeAdapter{
			owner:            "k8s-blueprint-operator",
			globalConfigRepo: repo,
		}

		// when
		err := sut.Deactivate(testCtx)

		// then
		require.NoError(t, err)
	})

	t.Run("should deactivate if no conflict occurs", func(t *testing.T) {
		// given
		mConfigRepo := newMockGeneralConfigRepository(t)
		globalConfig := config.CreateConfig(config.Entries{"maintenance": "{\"title\": \"title\", \"text\": \"text\", \"holder\": \"k8s-blueprint-operator\"}"})
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(globalConfig, nil)
		mConfigRepo.EXPECT().update(testCtx, configName("global-config"), config.SimpleDoguName(""), mock.Anything).Return(config.Config{}, nil)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		sut := &MaintenanceModeAdapter{
			owner:            "k8s-blueprint-operator",
			globalConfigRepo: repo,
		}

		// when
		err := sut.Deactivate(testCtx)

		// then
		require.NoError(t, err)
		_, ok := globalConfig.Get("maintenance")
		assert.False(t, ok)
	})

	t.Run("should return error on error updating config", func(t *testing.T) {
		// given
		mConfigRepo := newMockGeneralConfigRepository(t)
		globalConfig := config.CreateConfig(config.Entries{"maintenance": "{\"title\": \"title\", \"text\": \"text\", \"holder\": \"k8s-blueprint-operator\"}"})
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(globalConfig, nil)
		mConfigRepo.EXPECT().update(testCtx, configName("global-config"), config.SimpleDoguName(""), mock.Anything).Return(config.Config{}, assert.AnError)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		sut := &MaintenanceModeAdapter{
			owner:            "k8s-blueprint-operator",
			globalConfigRepo: repo,
		}

		// when
		err := sut.Deactivate(testCtx)

		// then
		require.Error(t, err)
		assert.ErrorContains(t, err, "could not update global config-map for activating maintenance mode")
		assert.True(t, errors.IsGenericError(err))
	})

	t.Run("should return error if another component holds the maintenance mode", func(t *testing.T) {
		// given
		mConfigRepo := newMockGeneralConfigRepository(t)
		globalConfig := config.CreateConfig(config.Entries{"maintenance": "{\"title\": \"title\", \"text\": \"text\", \"holder\": \"k8s-backup-operator\"}"})
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(globalConfig, nil)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		sut := &MaintenanceModeAdapter{
			owner:            "k8s-blueprint-operator",
			globalConfigRepo: repo,
		}

		// when
		err := sut.Deactivate(testCtx)

		// then
		require.Error(t, err)
		assert.ErrorContains(t, err, "maintenance mode {\"title\": \"title\", \"text\": \"text\", \"holder\": \"k8s-backup-operator\"} is already activated by another owner: k8s-backup-operator")
		assert.True(t, errors.IsConflictError(err))
	})
}

func TestNewMaintenanceModeAdapter(t *testing.T) {
	t.Run("should succeed", func(t *testing.T) {
		// when
		owner := "k8s-blueprint-operator"
		adapter := NewMaintenanceModeAdapter(owner, nil)

		// then
		assert.Equal(t, owner, adapter.owner)
	})
}
