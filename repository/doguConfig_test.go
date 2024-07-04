package repository

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

const _DoguName = config.SimpleDoguName("test")

func TestNewDoguConfigRepository(t *testing.T) {
	mClient := NewMockConfigMapClient(t)
	repo := NewDoguConfigRepository(mClient)
	assert.NotNil(t, repo)
}

func TestDoguConfigRepository_Get(t *testing.T) {
	t.Run("Get Dogu Config", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_DoguName.String())).Return(config.Config{PersistenceContext: resourceVersion}, nil)

		repo := &DoguConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		cfg, err := repo.Get(context.TODO(), _DoguName)
		assert.NoError(t, err)
		assert.Equal(t, _DoguName, cfg.DoguName)
		assert.Equal(t, resourceVersion, cfg.PersistenceContext)
	})

	t.Run("Config repo error", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().get(mock.Anything, createConfigName(_DoguName.String())).Return(config.Config{}, assert.AnError)

		repo := &DoguConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		_, err := repo.Get(context.TODO(), _DoguName)
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestDoguConfigRepository_Create(t *testing.T) {
	t.Run("Create Dogu Config", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().create(mock.Anything, createConfigName(_DoguName.String()), _DoguName, mock.Anything).Return(config.Config{PersistenceContext: resourceVersion}, nil)

		repo := &DoguConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		cfg, err := repo.Create(context.TODO(), config.CreateDoguConfig(_DoguName, make(config.Entries)))
		assert.NoError(t, err)
		assert.Equal(t, _DoguName, cfg.DoguName)
		assert.Equal(t, resourceVersion, cfg.PersistenceContext)
	})

	t.Run("Config repo error", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().create(mock.Anything, createConfigName(_DoguName.String()), _DoguName, mock.Anything).Return(config.Config{}, assert.AnError)

		repo := &DoguConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		_, err := repo.Create(context.TODO(), config.CreateDoguConfig(_DoguName, make(config.Entries)))
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestDoguConfigRepository_Update(t *testing.T) {
	t.Run("Update Dogu Config", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().update(mock.Anything, createConfigName(_DoguName.String()), _DoguName, mock.Anything).Return(config.Config{PersistenceContext: resourceVersion}, nil)

		repo := &DoguConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		cfg, err := repo.Update(context.TODO(), config.CreateDoguConfig(_DoguName, make(config.Entries)))
		assert.NoError(t, err)
		assert.Equal(t, _DoguName, cfg.DoguName)
		assert.Equal(t, resourceVersion, cfg.PersistenceContext)
	})

	t.Run("Config repo error", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().update(mock.Anything, createConfigName(_DoguName.String()), _DoguName, mock.Anything).Return(config.Config{}, assert.AnError)

		repo := &DoguConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		_, err := repo.Update(context.TODO(), config.CreateDoguConfig(_DoguName, make(config.Entries)))
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestDoguConfigRepository_SaveOrMerge(t *testing.T) {
	t.Run("Save&Merge Dogu Config", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().saveOrMerge(mock.Anything, createConfigName(_DoguName.String()), mock.Anything).Return(config.Config{PersistenceContext: resourceVersion}, nil)

		repo := &DoguConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		cfg, err := repo.SaveOrMerge(context.TODO(), config.CreateDoguConfig(_DoguName, make(config.Entries)))
		assert.NoError(t, err)
		assert.Equal(t, _DoguName, cfg.DoguName)
		assert.Equal(t, resourceVersion, cfg.PersistenceContext)
	})

	t.Run("Config repo error", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().saveOrMerge(mock.Anything, createConfigName(_DoguName.String()), mock.Anything).Return(config.Config{}, assert.AnError)

		repo := &DoguConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		_, err := repo.SaveOrMerge(context.TODO(), config.CreateDoguConfig(_DoguName, make(config.Entries)))
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestDoguConfigRepository_Delete(t *testing.T) {
	t.Run("Delete Dogu Config", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().delete(mock.Anything, createConfigName(_DoguName.String())).Return(nil)

		repo := &DoguConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		err := repo.Delete(context.TODO(), _DoguName)
		assert.NoError(t, err)
	})

	t.Run("Config repo error", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().delete(mock.Anything, createConfigName(_DoguName.String())).Return(assert.AnError)

		repo := &DoguConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		err := repo.Delete(context.TODO(), _DoguName)
		assert.ErrorIs(t, err, assert.AnError)
	})
}
