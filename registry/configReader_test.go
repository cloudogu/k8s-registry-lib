package registry

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_configReader_Exists(t *testing.T) {
	ctx := context.Background()

	t.Run("should check config exists", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"foo/bar": "value1"})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)

		cr := configReader{mockRepo}
		exists, err := cr.Exists(ctx, "foo/bar")

		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should check config not exists", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"foo/bar": "value1"})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)

		cr := configReader{mockRepo}
		exists, err := cr.Exists(ctx, "not/exists")

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("should fail to check if config exists on error in repo", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"foo/bar": "value1"})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, assert.AnError)

		cr := configReader{mockRepo}
		_, err := cr.Exists(ctx, "foo/bar")

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "could not read dogu config:")
	})
}

func Test_configReader_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("should get config", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"foo/bar": "value1"})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)

		cr := configReader{mockRepo}
		val, err := cr.Get(ctx, "foo/bar")

		require.NoError(t, err)
		assert.Equal(t, "value1", val)
	})

	t.Run("should fail to get config that does not exist", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"foo/bar": "value1"})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)

		cr := configReader{mockRepo}
		_, err := cr.Get(ctx, "not/exists")

		require.Error(t, err)
		assert.ErrorContains(t, err, "value for not/exists does not exist")
	})

	t.Run("should fail to get config on error in repo", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"foo/bar": "value1"})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, assert.AnError)

		cr := configReader{mockRepo}
		_, err := cr.Get(ctx, "foo/bar")

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "could not read dogu config:")
	})
}

func Test_configReader_GetAll(t *testing.T) {
	ctx := context.Background()

	t.Run("should get all config", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"foo/bar": "value1", "key1": "other"})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)

		cr := configReader{mockRepo}
		values, err := cr.GetAll(ctx)

		require.NoError(t, err)
		assert.Equal(t, map[string]string{"foo/bar": "value1", "key1": "other"}, values)
	})

	t.Run("should get empty config", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)

		cr := configReader{mockRepo}
		values, err := cr.GetAll(ctx)

		require.NoError(t, err)
		assert.Equal(t, map[string]string{}, values)
	})

	t.Run("should fail to get all config on error in repo", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"foo/bar": "value1"})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, assert.AnError)

		cr := configReader{mockRepo}
		_, err := cr.GetAll(ctx)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "could not read dogu config:")
	})
}
