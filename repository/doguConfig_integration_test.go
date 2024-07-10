//go:build integration
// +build integration

package repository

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"strings"
	"sync"
	"testing"
	"time"
)

func evaluateDoguConfig(ctx context.Context, t *testing.T, doguName config.SimpleDoguName, expected config.Config, cmClient corev1client.ConfigMapInterface) {
	cm, err := cmClient.Get(ctx, createConfigName(string(doguName)).String(), metav1.GetOptions{})
	assert.NoError(t, err)

	converter := &config.YamlConverter{}
	cmConfig, err := converter.Read(strings.NewReader(cm.Data["config.yaml"]))
	assert.NoError(t, err)
	assert.Equal(t, expected, config.CreateConfig(cmConfig))
}

func evaluateSensitiveDoguConfig(ctx context.Context, t *testing.T, doguName config.SimpleDoguName, expected config.Config, secretClient corev1client.SecretInterface) {
	cm, err := secretClient.Get(ctx, createConfigName(string(doguName)).String(), metav1.GetOptions{})
	assert.NoError(t, err)

	converter := &config.YamlConverter{}
	cmConfig, err := converter.Read(strings.NewReader(string(cm.Data["config.yaml"])))
	assert.NoError(t, err)
	assert.Equal(t, expected, config.CreateConfig(cmConfig))
}

func TestDoguConfigRepository(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	client, cleanUp := startContainerAndGetClient(ctx, t)
	defer cleanUp()

	cmClient := client.CoreV1().ConfigMaps(namespace)
	secretsClient := client.CoreV1().Secrets(namespace)

	var doguName config.SimpleDoguName = "myDogu"

	t.Run("test dogu-config-repo", func(t *testing.T) {
		repo := NewDoguConfigRepository(cmClient)

		//cleanUp
		_ = repo.Delete(ctx, doguName)

		cfg := config.DoguConfig{
			DoguName: doguName,
			Config: config.CreateConfig(config.Entries{
				"root":      "rootVal",
				"foo/sub/1": "val1",
				"foo/sub/2": "val2",
				"foo/sub/3": "val3",
				"foo/other": "otherVale",
			}),
		}
		cfg, err := repo.Create(ctx, cfg)
		assert.NoError(t, err)
		evaluateDoguConfig(ctx, t, doguName,
			config.CreateConfig(config.Entries{
				"root":      "rootVal",
				"foo/sub/1": "val1",
				"foo/sub/2": "val2",
				"foo/sub/3": "val3",
				"foo/other": "otherVale",
			}), cmClient)

		cfg.Config = cfg.Delete("foo/sub/2")

		cfg, err = repo.Update(ctx, cfg)
		assert.NoError(t, err)
		evaluateDoguConfig(ctx, t, doguName,
			config.CreateConfig(config.Entries{
				"root":      "rootVal",
				"foo/sub/1": "val1",
				"foo/sub/3": "val3",
				"foo/other": "otherVale",
			}), cmClient)
	})

	t.Run("test sensitive-dogu-config-repo", func(t *testing.T) {
		repo := NewSensitiveDoguConfigRepository(secretsClient)

		var doguName config.SimpleDoguName = "myDogu"

		//cleanUp
		_ = repo.Delete(ctx, doguName)

		cfg := config.DoguConfig{
			DoguName: doguName,
			Config: config.CreateConfig(config.Entries{
				"root":      "rootVal",
				"foo/sub/1": "val1",
				"foo/sub/2": "val2",
				"foo/sub/3": "val3",
				"foo/other": "otherVale",
			}),
		}
		cfg, err := repo.Create(ctx, cfg)
		assert.NoError(t, err)
		evaluateSensitiveDoguConfig(ctx, t, doguName,
			config.CreateConfig(config.Entries{
				"root":      "rootVal",
				"foo/sub/1": "val1",
				"foo/sub/2": "val2",
				"foo/sub/3": "val3",
				"foo/other": "otherVale",
			}), secretsClient)

		cfg.Config = cfg.Delete("foo/sub/2")

		cfg, err = repo.Update(ctx, cfg)
		assert.NoError(t, err)
		evaluateSensitiveDoguConfig(ctx, t, doguName,
			config.CreateConfig(config.Entries{
				"root":      "rootVal",
				"foo/sub/1": "val1",
				"foo/sub/3": "val3",
				"foo/other": "otherVale",
			}), secretsClient)
	})
}

func TestDoguConfigWatch(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	client, cleanUp := startContainerAndGetClient(ctx, t)
	defer cleanUp()

	cmClient := client.CoreV1().ConfigMaps(namespace)
	secretsClient := client.CoreV1().Secrets(namespace)

	var doguName config.SimpleDoguName = "myDogu"

	t.Run("Watch dogu config", func(t *testing.T) {
		doguWatchCtx, gCancel := context.WithTimeout(ctx, 3*time.Second)
		defer gCancel()

		repo := NewDoguConfigRepository(cmClient)

		//cleanUp
		_ = repo.Delete(ctx, doguName)

		cfg := config.DoguConfig{
			DoguName: doguName,
			Config: config.CreateConfig(config.Entries{
				"root":      "rootVal",
				"foo/sub/1": "val1",
				"foo/sub/2": "val2",
				"foo/sub/3": "val3",
				"foo/other": "otherVale",
			}),
		}
		cfg, err := repo.Create(ctx, cfg)
		assert.NoError(t, err)

		resultChan, err := repo.Watch(doguWatchCtx, doguName, config.KeyFilter("foo/sub/3"))
		assert.NoError(t, err)

		var wg sync.WaitGroup
		waitCh := make(chan struct{})

		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-doguWatchCtx.Done():
				t.Errorf("context reach end before receiving watch result")
			case change := <-resultChan:
				assert.NoError(t, change.Err)

				assert.NotEqual(t, change.PrevState.PersistenceContext, change.NewState.PersistenceContext)

				assert.Empty(t, change.PrevState.Diff(config.CreateConfig(config.Entries{
					"root":      "rootVal",
					"foo/sub/1": "val1",
					"foo/sub/2": "val2",
					"foo/sub/3": "val3",
					"foo/other": "otherVale",
				})))

				assert.Empty(t, change.NewState.Diff(config.CreateConfig(config.Entries{
					"root":      "rootVal",
					"foo/sub/1": "val1",
					"foo/sub/2": "val2",
					"foo/sub/3": "newVal",
					"foo/other": "otherVale",
				})))
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			cfg.Config, err = cfg.Set("foo/sub/3", "newVal")
			require.NoError(t, err)
			cfg, err = repo.Update(doguWatchCtx, cfg)
			require.NoError(t, err, "could not set new value for watchKey")
		}()

		go func() {
			wg.Wait()
			close(waitCh)
		}()

		select {
		case <-doguWatchCtx.Done():
			t.Errorf("context reach end before receiving watch result")
		case <-waitCh:
			t.Log("Finished dogu config watch without errors")
		}
	})

	t.Run("Watch sensitive dogu config", func(t *testing.T) {
		doguWatchCtx, gCancel := context.WithTimeout(ctx, 3*time.Second)
		defer gCancel()

		repo := NewSensitiveDoguConfigRepository(secretsClient)

		//cleanUp
		_ = repo.Delete(ctx, doguName)

		cfg := config.DoguConfig{
			DoguName: doguName,
			Config: config.CreateConfig(config.Entries{
				"root":      "rootVal",
				"foo/sub/1": "val1",
				"foo/sub/2": "val2",
				"foo/sub/3": "val3",
				"foo/other": "otherVale",
			}),
		}
		cfg, err := repo.Create(ctx, cfg)
		assert.NoError(t, err)

		resultChan, err := repo.Watch(doguWatchCtx, doguName, config.KeyFilter("foo/sub/3"))
		assert.NoError(t, err)

		var wg sync.WaitGroup
		waitCh := make(chan struct{})

		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-doguWatchCtx.Done():
				t.Errorf("context reach end before receiving watch result")
			case change := <-resultChan:
				assert.NoError(t, change.Err)

				assert.NotEqual(t, change.PrevState.PersistenceContext, change.NewState.PersistenceContext)

				assert.Empty(t, change.PrevState.Diff(config.CreateConfig(config.Entries{
					"root":      "rootVal",
					"foo/sub/1": "val1",
					"foo/sub/2": "val2",
					"foo/sub/3": "val3",
					"foo/other": "otherVale",
				})))

				assert.Empty(t, change.NewState.Diff(config.CreateConfig(config.Entries{
					"root":      "rootVal",
					"foo/sub/1": "val1",
					"foo/sub/2": "val2",
					"foo/sub/3": "newVal",
					"foo/other": "otherVale",
				})))
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			cfg.Config, err = cfg.Set("foo/sub/3", "newVal")
			require.NoError(t, err)
			cfg, err = repo.Update(doguWatchCtx, cfg)
			require.NoError(t, err, "could not set new value for watchKey")
		}()

		go func() {
			wg.Wait()
			close(waitCh)
		}()

		select {
		case <-doguWatchCtx.Done():
			t.Errorf("context reach end before receiving watch result")
		case <-waitCh:
			t.Log("Finished sensitive dogu config watch without errors")
		}
	})
}
