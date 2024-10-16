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

func Test_defaultSwitcher_Activate(t *testing.T) {
	t.Run("should fail to activate maintenance mode", func(t *testing.T) {
		// given
		testctx := context.Background()
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(config.Config{}, assert.AnError)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		sut := &MaintenanceModeAdapter{
			owner:            "testOwner",
			globalConfigRepo: repo,
		}

		// when
		err := sut.Activate(testctx, MaintenanceModeDescription{
			Title: "myTitle",
			Text:  "myText",
		})

		// then
		require.Error(t, err)
		assert.True(t, errors.IsGenericError(err))
		assert.ErrorContains(t, err, "could not get contents of global config-map for activating maintenance mode")
	})
	t.Run("should succeed to activate maintenance mode", func(t *testing.T) {
		// given
		//expectedJson := `{"title":"myTitle","text":"myText","holder":"k8s-blueprint-operator"}`
		//globalConfigMock := newMockGlobalConfig(t)
		//globalConfigMock.EXPECT().Set("maintenance", expectedJson).Return(nil)
		//
		//sut := &defaultSwitcher{globalConfig: globalConfigMock}
		//
		//// when
		//err := sut.activate(domainservice.MaintenancePageModel{
		//	Title: "myTitle",
		//	Text:  "myText",
		//})
		//
		//// then
		//require.NoError(t, err)
	})
}

func TestSwitch_Deactivate(t *testing.T) {
	t.Run("should fail to deactivate maintenance mode", func(t *testing.T) {
		// given
		//configMapClientMock := NewMockConfigMapClient(t)
		//configMapClientMock.EXPECT().Delete("maintenance").Return(assert.AnError)
		//
		//sut := NewMaintenanceModeAdapter("testOwner", configMapClientMock)
		//
		//// when
		//err := sut.Deactivate(context.Background())
		//
		//// then
		//require.Error(t, err)
		//assert.ErrorIs(t, err, assert.AnError)
		//assert.ErrorContains(t, err, "failed to delete maintenance mode registry key")
	})
	t.Run("should succeed to deactivate maintenance mode", func(t *testing.T) {
		// given
		//globalConfigMock := newMockGlobalConfig(t)
		//globalConfigMock.EXPECT().Delete("maintenance").Return(nil)
		//
		//sut := &defaultSwitcher{globalConfig: globalConfigMock}
		//
		//// when
		//err := sut.deactivate()
		//
		//// then
		//require.NoError(t, err)
	})
}
