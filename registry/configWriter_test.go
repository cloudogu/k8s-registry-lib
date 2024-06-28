package registry

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_configWriter_Set(t *testing.T) {
	ctx := context.Background()

	t.Run("should set config", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"value/key": "value1"})
		expectedConf := config.CreateConfig(config.Entries{"value/key": "value1"})
		expectedConf.Set("foo/bar", "myVal")

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)
		mockRepo.EXPECT().write(ctx, expectedConf).Return(nil)

		cw := configWriter{mockRepo}
		err := cw.Set(ctx, "foo/bar", "myVal")

		require.NoError(t, err)
	})

	t.Run("error while setting config", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"value/key": "value1"})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)

		cw := configWriter{mockRepo}
		err := cw.Set(ctx, "value", "myVal")

		assert.Error(t, err)
	})

	t.Run("should create new config if config not found", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{})
		expectedConf := config.CreateConfig(config.Entries{})
		expectedConf.Set("foo/bar", "myVal")

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, ErrConfigNotFound)
		mockRepo.EXPECT().write(ctx, expectedConf).Return(nil)

		cw := configWriter{mockRepo}
		err := cw.Set(ctx, "foo/bar", "myVal")

		require.NoError(t, err)
	})

	t.Run("should fail to set config on get-error in repo", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"foo/bar": "value1"})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, assert.AnError)

		cw := configWriter{mockRepo}
		err := cw.Set(ctx, "foo/bar", "myVal")

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "could not read config:")
	})

	t.Run("should fail to set config on write-error in repo", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"value/key": "value1"})
		expectedConf := config.CreateConfig(config.Entries{"value/key": "value1"})
		expectedConf.Set("foo/bar", "myVal")

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)
		mockRepo.EXPECT().write(ctx, expectedConf).Return(assert.AnError)

		cw := configWriter{mockRepo}
		err := cw.Set(ctx, "foo/bar", "myVal")

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "could not write config after updating value:")
	})
}

func Test_configWriter_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete config", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"value/key": "value1", "foo/bar": "myValue"})
		expectedConf := config.CreateConfig(config.Entries{"value/key": "value1", "foo/bar": "myValue"})
		expectedConf.Delete("foo/bar")

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)
		mockRepo.EXPECT().write(ctx, expectedConf).Return(nil)

		cw := configWriter{mockRepo}
		err := cw.Delete(ctx, "foo/bar")

		require.NoError(t, err)
	})

	t.Run("should fail to delete config on get-error in repo", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"foo/bar": "value1"})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, assert.AnError)

		cw := configWriter{mockRepo}
		err := cw.Delete(ctx, "foo/bar")

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "could not read config:")
	})

	t.Run("should fail to delete config on write-error in repo", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"value/key": "value1", "foo/bar": "myValue"})
		expectedConf := config.CreateConfig(config.Entries{"value/key": "value1", "foo/bar": "myValue"})
		expectedConf.Delete("foo/bar")

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)
		mockRepo.EXPECT().write(ctx, expectedConf).Return(assert.AnError)

		cw := configWriter{mockRepo}
		err := cw.Delete(ctx, "foo/bar")

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "could not write config after deleting key foo/bar:")
	})
}

func Test_configWriter_DeleteRecursive(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete config recursive", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"value/key": "value1", "foo/bar": "myValue"})
		expectedConf := config.CreateConfig(config.Entries{"value/key": "value1", "foo/bar": "myValue"})
		expectedConf.DeleteRecursive("foo")

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)
		mockRepo.EXPECT().write(ctx, expectedConf).Return(nil)

		cw := configWriter{mockRepo}
		err := cw.DeleteRecursive(ctx, "foo")

		require.NoError(t, err)
	})

	t.Run("should fail to delete config recursive on get-error in repo", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"value/key": "value1", "foo/bar": "myValue"})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, assert.AnError)

		cw := configWriter{mockRepo}
		err := cw.DeleteRecursive(ctx, "foo")

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "could not read config:")
	})

	t.Run("should fail to delete config recursive on write-error in repo", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"value/key": "value1", "foo/bar": "myValue"})
		expectedConf := config.CreateConfig(config.Entries{"value/key": "value1", "foo/bar": "myValue"})
		expectedConf.DeleteRecursive("foo")

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)
		mockRepo.EXPECT().write(ctx, expectedConf).Return(assert.AnError)

		cw := configWriter{mockRepo}
		err := cw.DeleteRecursive(ctx, "foo")

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "could not write config after recursively deleting key foo:")
	})
}

func Test_configWriter_DeleteAll(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete all config", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"value/key": "value1"})
		expectedConf := config.CreateConfig(config.Entries{"value/key": "value1"})
		expectedConf.DeleteAll()

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)
		mockRepo.EXPECT().write(ctx, expectedConf).Return(nil)

		cw := configWriter{mockRepo}
		err := cw.DeleteAll(ctx)

		require.NoError(t, err)
	})

	t.Run("should fail to delete all config on get-error in repo", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"value/key": "value1"})

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, assert.AnError)

		cw := configWriter{mockRepo}
		err := cw.DeleteAll(ctx)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "could not read dogu config:")
	})

	t.Run("should fail to delete all config on write-error in repo", func(t *testing.T) {
		conf := config.CreateConfig(config.Entries{"value/key": "value1"})
		expectedConf := config.CreateConfig(config.Entries{"value/key": "value1"})
		expectedConf.DeleteAll()

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().get(ctx).Return(conf, nil)
		mockRepo.EXPECT().write(ctx, expectedConf).Return(assert.AnError)

		cw := configWriter{mockRepo}
		err := cw.DeleteAll(ctx)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "could not write config after deleting all keys:")
	})
}
