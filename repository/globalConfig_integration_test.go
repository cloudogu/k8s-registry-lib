//go:build integration
// +build integration

package repository

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
	"sync"
	"testing"
	"time"
)

const namespace = "testing"

func startContainerAndGetClient(ctx context.Context, t *testing.T) (*kubernetes.Clientset, func()) {
	k3sContainer, err := k3s.RunContainer(ctx,
		testcontainers.WithImage("docker.io/rancher/k3s:v1.27.1-k3s1"),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container
	cleanUp := func() {
		if lErr := k3sContainer.Terminate(ctx); lErr != nil {
			t.Fatal(lErr)
		}
	}

	kubeConfigYaml, err := k3sContainer.GetKubeConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}

	restcfg, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigYaml)
	if err != nil {
		t.Fatal(err)
	}

	k8s, err := kubernetes.NewForConfig(restcfg)
	if err != nil {
		t.Fatal(err)
	}

	_, err = k8s.CoreV1().Namespaces().Create(ctx, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	return k8s, cleanUp
}

func evaluateConfig(ctx context.Context, t *testing.T, expected config.Config, cmClient corev1client.ConfigMapInterface) {
	cm, err := cmClient.Get(ctx, "global-config", metav1.GetOptions{})
	assert.NoError(t, err)

	converter := &config.YamlConverter{}
	cmConfig, err := converter.Read(strings.NewReader(cm.Data["config.yaml"]))
	assert.NoError(t, err)
	assert.Equal(t, expected, config.CreateConfig(cmConfig))
}

func TestGlobalConfigRepository(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	client, cleanUp := startContainerAndGetClient(ctx, t)
	defer cleanUp()

	cmClient := client.CoreV1().ConfigMaps(namespace)

	repo := NewGlobalConfigRepository(cmClient)

	//cleanUp
	_ = repo.Delete(ctx)

	cfg := config.GlobalConfig{
		config.CreateConfig(config.Entries{
			"root":      "rootVal",
			"foo/sub/1": "val1",
			"foo/sub/2": "val2",
			"foo/sub/3": "val3",
			"foo/other": "otherVale",
		}),
	}
	cfg, err := repo.Create(ctx, cfg)
	assert.NoError(t, err)
	evaluateConfig(ctx, t,
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
	evaluateConfig(ctx, t,
		config.CreateConfig(config.Entries{
			"root":      "rootVal",
			"foo/sub/1": "val1",
			"foo/sub/3": "val3",
			"foo/other": "otherVale",
		}), cmClient)
}

func TestGlobalConfigWatch(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	client, cleanUp := startContainerAndGetClient(ctx, t)
	defer cleanUp()

	cmClient := client.CoreV1().ConfigMaps(namespace)

	t.Run("Watch global config", func(t *testing.T) {
		globalWatchCtx, gCancel := context.WithTimeout(ctx, 3*time.Second)
		defer gCancel()

		repo := NewGlobalConfigRepository(cmClient)

		//cleanUp
		_ = repo.Delete(ctx)

		cfg := config.GlobalConfig{
			config.CreateConfig(config.Entries{
				"root":      "rootVal",
				"foo/sub/1": "val1",
				"foo/sub/2": "val2",
				"foo/sub/3": "val3",
				"foo/other": "otherVale",
			}),
		}
		cfg, err := repo.Create(ctx, cfg)
		assert.NoError(t, err)

		resultChan, err := repo.Watch(globalWatchCtx, config.KeyFilter("foo/sub/3"))
		assert.NoError(t, err)

		var wg sync.WaitGroup
		waitCh := make(chan struct{})

		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-globalWatchCtx.Done():
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
			cfg, err = repo.Update(globalWatchCtx, cfg)
			require.NoError(t, err, "could not set new value for watchKey")
		}()

		go func() {
			wg.Wait()
			close(waitCh)
		}()

		select {
		case <-globalWatchCtx.Done():
			t.Errorf("context reach end before receiving watch result")
		case <-waitCh:
			t.Log("Finished global config watch without errors")
		}
	})
}
