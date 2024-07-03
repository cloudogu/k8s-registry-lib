package repository

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestNewGlobalConfigRepository(t *testing.T) {
	mClient := NewMockConfigMapClient(t)
	repo := NewGlobalConfigRepository(mClient)
	assert.NotNil(t, repo)
}

func TestGlobalConfigRepository_Get(t *testing.T) {
	t.Run("Get Global Config", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(config.Config{}, nil)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		_, err := repo.Get(context.TODO())
		assert.NoError(t, err)
	})

	t.Run("Config repo error", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(config.Config{}, assert.AnError)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		_, err := repo.Get(context.TODO())
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestGlobalConfigRepository_Create(t *testing.T) {
	t.Run("Create Global Config", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().create(mock.Anything, createConfigName(_SimpleGlobalConfigName), config.SimpleDoguName(""), mock.Anything).Return(config.Config{PersistenceContext: resourceVersion}, nil)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		gCfg, err := repo.Create(context.TODO(), config.CreateGlobalConfig(make(config.Entries)))
		assert.NoError(t, err)
		assert.Equal(t, resourceVersion, gCfg.PersistenceContext)
	})

	t.Run("Config repo error", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().create(mock.Anything, createConfigName(_SimpleGlobalConfigName), config.SimpleDoguName(""), mock.Anything).Return(config.Config{}, assert.AnError)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		_, err := repo.Create(context.TODO(), config.CreateGlobalConfig(make(config.Entries)))
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestGlobalConfigRepository_Update(t *testing.T) {
	t.Run("Update Global Config", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().update(mock.Anything, createConfigName(_SimpleGlobalConfigName), config.SimpleDoguName(""), mock.Anything).Return(config.Config{PersistenceContext: resourceVersion}, nil)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		gCfg, err := repo.Update(context.TODO(), config.CreateGlobalConfig(make(config.Entries)))
		assert.NoError(t, err)
		assert.Equal(t, resourceVersion, gCfg.PersistenceContext)
	})

	t.Run("Config repo error", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().update(mock.Anything, createConfigName(_SimpleGlobalConfigName), config.SimpleDoguName(""), mock.Anything).Return(config.Config{}, assert.AnError)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		_, err := repo.Update(context.TODO(), config.CreateGlobalConfig(make(config.Entries)))
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestGlobalConfigRepository_SaveOrMerge(t *testing.T) {
	t.Run("Save&Merge Global Config", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().saveOrMerge(mock.Anything, createConfigName(_SimpleGlobalConfigName), mock.Anything).Return(config.Config{PersistenceContext: resourceVersion}, nil)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		gCfg, err := repo.SaveOrMerge(context.TODO(), config.CreateGlobalConfig(make(config.Entries)))
		assert.NoError(t, err)
		assert.Equal(t, resourceVersion, gCfg.PersistenceContext)
	})

	t.Run("Config repo error", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().saveOrMerge(mock.Anything, createConfigName(_SimpleGlobalConfigName), mock.Anything).Return(config.Config{}, assert.AnError)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		_, err := repo.SaveOrMerge(context.TODO(), config.CreateGlobalConfig(make(config.Entries)))
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestGlobalConfigRepository_Delete(t *testing.T) {
	t.Run("Delete Global Config", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().delete(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(nil)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		err := repo.Delete(context.TODO())
		assert.NoError(t, err)
	})

	t.Run("Config repo error", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().delete(mock.Anything, createConfigName(_SimpleGlobalConfigName)).Return(assert.AnError)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		err := repo.Delete(context.TODO())
		assert.ErrorIs(t, err, assert.AnError)
	})
}
