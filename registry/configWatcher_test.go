package registry

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_compareConfigForSingleKey(t *testing.T) {
	tests := []struct {
		name      string
		oldConfig config.Config
		newConfig config.Config
		configKey string
		want      ConfigModification
		want1     bool
	}{
		{
			"should compare existing keys with different values",
			config.CreateConfig(config.Data{"foo": "bar"}),
			config.CreateConfig(config.Data{"foo": "value"}),
			"foo",
			ConfigModification{"bar", "value"},
			true,
		},
		{
			"should compare existing keys with same values",
			config.CreateConfig(config.Data{"foo": "bar"}),
			config.CreateConfig(config.Data{"foo": "bar"}),
			"foo",
			ConfigModification{"bar", "bar"},
			false,
		},
		{
			"should compare old-key exists, new key does not exist",
			config.CreateConfig(config.Data{"foo": "bar"}),
			config.CreateConfig(config.Data{"bar": "value"}),
			"foo",
			ConfigModification{"bar", ""},
			true,
		},
		{
			"should compare old-key does not exist, new key exists",
			config.CreateConfig(config.Data{"bar": "bar"}),
			config.CreateConfig(config.Data{"foo": "value"}),
			"foo",
			ConfigModification{"", "value"},
			true,
		},
		{
			"should compare both keys do not exist",
			config.CreateConfig(config.Data{"bar": "bar"}),
			config.CreateConfig(config.Data{"bar": "value"}),
			"foo",
			ConfigModification{"", ""},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := compareConfigForSingleKey(tt.oldConfig, tt.newConfig, tt.configKey)
			assert.Equalf(t, tt.want, got, "compareConfigForSingleKey(%v, %v, %v)", tt.oldConfig, tt.newConfig, tt.configKey)
			assert.Equalf(t, tt.want1, got1, "compareConfigForSingleKey(%v, %v, %v)", tt.oldConfig, tt.newConfig, tt.configKey)
		})
	}
}

func Test_compareConfigs(t *testing.T) {
	tests := []struct {
		name      string
		oldConfig config.Config
		newConfig config.Config
		configKey string
		recursive bool
		want      map[string]ConfigModification
	}{
		{
			"should compare configs non-recursive",
			config.CreateConfig(config.Data{"foo": "bar"}),
			config.CreateConfig(config.Data{"foo": "value"}),
			"foo",
			false,
			map[string]ConfigModification{"foo": {"bar", "value"}},
		},
		{
			"should compare configs non-recursive with same value",
			config.CreateConfig(config.Data{"foo": "bar"}),
			config.CreateConfig(config.Data{"foo": "bar"}),
			"foo",
			false,
			map[string]ConfigModification{},
		},
		{
			"should compare configs recursive",
			config.CreateConfig(config.Data{"foo/1": "bar", "foo/2": "bar2", "foo/sub/3": "bar3"}),
			config.CreateConfig(config.Data{"foo/1": "val", "foo/2": "val2", "foo/sub/3": "val3"}),
			"foo",
			true,
			map[string]ConfigModification{
				"foo/1":     {"bar", "val"},
				"foo/2":     {"bar2", "val2"},
				"foo/sub/3": {"bar3", "val3"},
			},
		},
		{
			"should compare configs recursive with non matching",
			config.CreateConfig(config.Data{"foo/1": "bar", "foo/2": "bar2", "foo/sub/3": "bar3"}),
			config.CreateConfig(config.Data{"foo/1": "val", "foo/2": "val2", "foo/sub/3": "val3"}),
			"bar",
			true,
			map[string]ConfigModification{},
		},
		{
			"should compare configs recursive with deleted and added keys",
			config.CreateConfig(config.Data{"foo/1": "bar", "foo/2": "bar2"}),
			config.CreateConfig(config.Data{"foo/1": "val", "foo/sub/3": "val3"}),
			"foo",
			true,
			map[string]ConfigModification{
				"foo/1":     {"bar", "val"},
				"foo/2":     {"bar2", ""},
				"foo/sub/3": {"", "val3"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, compareConfigs(tt.oldConfig, tt.newConfig, tt.configKey, tt.recursive), "compareConfigs(%v, %v, %v, %v)", tt.oldConfig, tt.newConfig, tt.configKey, tt.recursive)
		})
	}
}

func Test_configWatcher_Watch(t *testing.T) {
	ctx := context.Background()

	t.Run("should watch config", func(t *testing.T) {
		resultChan := make(chan configWatchResult)
		confWatch := &configWatch{
			ResultChan:    resultChan,
			InitialConfig: config.CreateConfig(config.Data{"foo": "bar"}),
		}

		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().watch(mock.Anything).Return(confWatch, nil)

		watcher := configWatcher{mockRepo}

		watch, err := watcher.Watch(ctx, "foo", false)

		require.NoError(t, err)
		assert.NotNil(t, watch.ResultChan)
		assert.NotNil(t, watch.cancelWatchCtx)

		cancel := make(chan bool, 1)

		go func() {
			resultChan <- configWatchResult{config.CreateConfig(config.Data{"foo": "val"}), nil}
			resultChan <- configWatchResult{config.CreateConfig(config.Data{"foo": "val"}), nil}
			resultChan <- configWatchResult{config.Config{}, assert.AnError}
		}()

		go func() {
			i := 0
			for result := range watch.ResultChan {
				if i == 0 {
					assert.NoError(t, result.Err)
					assert.Equal(t, map[string]ConfigModification{"foo": {"bar", "val"}}, result.ModifiedKeys)
				}

				if i == 1 {
					assert.Error(t, result.Err)
					assert.ErrorIs(t, result.Err, assert.AnError)
					assert.ErrorContains(t, result.Err, "error watching config for key foo:")
					cancel <- true
				}

				i++
			}
		}()

		select {
		case <-cancel:
			close(resultChan)
		case <-time.After(5 * time.Second):
			close(resultChan)
			t.Errorf("did not reach all evente in time")
		}
	})

	t.Run("should fail to watch config for error while starting watch", func(t *testing.T) {
		mockRepo := newMockConfigRepository(t)
		mockRepo.EXPECT().watch(mock.Anything).Return(nil, assert.AnError)

		watcher := configWatcher{mockRepo}

		_, err := watcher.Watch(ctx, "foo", false)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "could not watch config:")
	})

}

func TestConfigWatch_Stop(t *testing.T) {
	t.Run("should call context-cancel when stop is called", func(t *testing.T) {
		ctx, cancelFunc := context.WithCancel(context.Background())

		watcher := ConfigWatch{nil, cancelFunc}

		watcher.Stop()

		require.Error(t, ctx.Err())
		assert.ErrorIs(t, ctx.Err(), context.Canceled)
	})
}
