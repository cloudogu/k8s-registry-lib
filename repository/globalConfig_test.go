package repository

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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

func TestGlobalConfigRepository_Watch(t *testing.T) {
	ctx := context.Background()

	t.Run("should watch config with filters", func(t *testing.T) {
		mockResultChan := make(chan configWatchResult)

		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().watch(ctx, createConfigName(_SimpleGlobalConfigName), mock.AnythingOfType("config.WatchFilter")).Return(mockResultChan, nil)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		resultChan, err := repo.Watch(ctx, config.KeyFilter("foo"))
		require.NoError(t, err)
		require.NotNil(t, resultChan)

		cancel := make(chan bool, 1)

		go func() {
			mockResultChan <- configWatchResult{config.CreateConfig(config.Entries{"foo": "val"}), config.CreateConfig(config.Entries{"foo": "val2"}), nil}
			mockResultChan <- configWatchResult{config.CreateConfig(config.Entries{"foo": "val2"}), config.CreateConfig(nil), nil}
			mockResultChan <- configWatchResult{config.CreateConfig(nil), config.CreateConfig(nil), assert.AnError}
		}()

		go func() {
			i := 0
			for result := range resultChan {
				if i == 0 {
					assert.NoError(t, result.Err)
					assert.Equal(t, result.PrevState, config.GlobalConfig{Config: config.CreateConfig(config.Entries{"foo": "val"})})
					assert.Equal(t, result.NewState, config.GlobalConfig{Config: config.CreateConfig(config.Entries{"foo": "val2"})})
				}

				if i == 1 {
					assert.NoError(t, result.Err)
					assert.Equal(t, result.PrevState, config.GlobalConfig{Config: config.CreateConfig(config.Entries{"foo": "val2"})})
					assert.Equal(t, result.NewState, config.GlobalConfig{Config: config.CreateConfig(nil)})
				}

				if i == 2 {
					assert.Error(t, result.Err)
					assert.ErrorIs(t, result.Err, assert.AnError)
					cancel <- true
				}

				i++
			}
		}()

		select {
		case <-cancel:
			close(mockResultChan)
		case <-time.After(5 * time.Second):
			close(mockResultChan)
			t.Errorf("did not reach all evente in time")
		}
	})

	t.Run("should fail to watch config for error while starting watch", func(t *testing.T) {
		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().watch(ctx, createConfigName(_SimpleGlobalConfigName)).Return(nil, assert.AnError)

		repo := &GlobalConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		_, err := repo.Watch(ctx)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "unable to start watch for global config:")
	})
}
