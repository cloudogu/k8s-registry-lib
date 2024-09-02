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

const _DoguName = config.SimpleDoguName("test")

func TestNewDoguConfigRepository(t *testing.T) {
	mClient := NewMockConfigMapClient(t)
	mInformer := NewMockConfigMapInformer(t)
	repo := NewDoguConfigRepository(mClient, mInformer)
	assert.NotNil(t, repo)
	assert.Equal(t, mClient, repo.generalConfigRepository.(configRepository).client.(configMapClient).client)
}

func TestNewSensitiveDoguConfigRepository(t *testing.T) {
	sClient := NewMockSecretClient(t)
	repo := NewSensitiveDoguConfigRepository(sClient)
	assert.NotNil(t, repo)
	assert.Equal(t, sClient, repo.generalConfigRepository.(configRepository).client.(secretClient).client)
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

func TestDoguConfigRepository_Watch(t *testing.T) {
	ctx := context.Background()

	t.Run("should watch config with filters", func(t *testing.T) {
		mockResultChan := make(chan configWatchResult)

		mConfigRepo := newMockGeneralConfigRepository(t)
		mConfigRepo.EXPECT().watch(ctx, createConfigName("myDogu"), mock.AnythingOfType("config.WatchFilter")).Return(mockResultChan, nil)

		repo := &DoguConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		resultChan, err := repo.Watch(ctx, "myDogu", config.KeyFilter("foo"))
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
					assert.Equal(t, result.PrevState, config.DoguConfig{DoguName: "myDogu", Config: config.CreateConfig(config.Entries{"foo": "val"})})
					assert.Equal(t, result.NewState, config.DoguConfig{DoguName: "myDogu", Config: config.CreateConfig(config.Entries{"foo": "val2"})})
				}

				if i == 1 {
					assert.NoError(t, result.Err)
					assert.Equal(t, result.PrevState, config.DoguConfig{DoguName: "myDogu", Config: config.CreateConfig(config.Entries{"foo": "val2"})})
					assert.Equal(t, result.NewState, config.DoguConfig{DoguName: "myDogu", Config: config.CreateConfig(nil)})
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
		mConfigRepo.EXPECT().watch(ctx, createConfigName("myDogu")).Return(nil, assert.AnError)

		repo := &DoguConfigRepository{
			generalConfigRepository: mConfigRepo,
		}

		_, err := repo.Watch(ctx, "myDogu")

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "unable to start watch for config from dogu myDogu:")
	})
}
