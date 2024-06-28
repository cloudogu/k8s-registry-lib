//go:build integration
// +build integration

package registry

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

func evaluateKeyValue(ctx context.Context, t *testing.T, key string, expected string, expectExist bool, reader ConfigurationReader) {
	exists, err := reader.Exists(ctx, key)
	assert.NoError(t, err, fmt.Sprintf("failed to check if exists: %s", err))

	assert.Equal(t, expectExist, exists, fmt.Sprintf("exists expected %v but got %v", expectExist, exists))

	if !expectExist {
		return
	}

	value, err := reader.Get(ctx, key)
	assert.NoError(t, err, "failed to get key", key)
	assert.Equal(t, expected, value)
}

func testKey(ctx context.Context, t *testing.T, keyname string, registry ConfigurationRegistry) {
	key := fmt.Sprintf("%s/a", keyname)
	err := registry.Set(ctx, key, "value")
	assert.NoError(t, err, "failed to set key", key)
	evaluateKeyValue(ctx, t, key, "value", true, registry)

	err = registry.Delete(ctx, key)
	assert.NoError(t, err, "failed to delete key", key)
	evaluateKeyValue(ctx, t, key, "", false, registry)

	key = fmt.Sprintf("%s/b/c", keyname)
	err = registry.Set(ctx, key, "value")
	assert.NoError(t, err, "failed to set key", key)
	evaluateKeyValue(ctx, t, key, "value", true, registry)

	err = registry.DeleteRecursive(ctx, keyname)
	assert.NoError(t, err, "failed to delete key", key)
	evaluateKeyValue(ctx, t, keyname, "", false, registry)
	evaluateKeyValue(ctx, t, key, "", false, registry)
	evaluateKeyValue(ctx, t, fmt.Sprintf("%s/a", keyname), "", false, registry)
	evaluateKeyValue(ctx, t, fmt.Sprintf("%s/b", keyname), "", false, registry)

	subkey1 := fmt.Sprintf("%s/b/c", keyname)
	err = registry.Set(ctx, subkey1, "value")
	assert.NoError(t, err, "failed to set key", subkey1)
	evaluateKeyValue(ctx, t, subkey1, "value", true, registry)

	subkey2 := fmt.Sprintf("%s/b/d", keyname)
	err = registry.Set(ctx, subkey2, "value")
	assert.NoError(t, err, "failed to set key", subkey2)
	evaluateKeyValue(ctx, t, subkey2, "value", true, registry)

	err = registry.DeleteRecursive(ctx, fmt.Sprintf("%s/b/", keyname))
	assert.NoError(t, err, "failed to delete recursive key", fmt.Sprintf("%s/b/", keyname))
	evaluateKeyValue(ctx, t, subkey1, "", false, registry)
	evaluateKeyValue(ctx, t, subkey2, "", false, registry)

	err = registry.Set(ctx, fmt.Sprintf("%s/a", keyname), "value")
	require.NoError(t, err)
	err = registry.Set(ctx, fmt.Sprintf("%s/b", keyname), "value")
	require.NoError(t, err)
	err = registry.Set(ctx, fmt.Sprintf("%s/c", keyname), "value")
	require.NoError(t, err)
	err = registry.Set(ctx, fmt.Sprintf("%s/d/e", keyname), "value")
	require.NoError(t, err)

	err = registry.DeleteAll(ctx)
	assert.NoError(t, err, "failed to delete all")

	cpReg, err := registry.GetAll(ctx)
	assert.NoError(t, err, "failed to get all")
	assert.True(t, len(cpReg) == 0)
}

func testAllFunctions(ctx context.Context, t *testing.T, registry ConfigurationRegistry, keyname string) {
	err := registry.DeleteAll(ctx)
	require.NoError(t, err, "clean up for testAllFunctions failed")

	regData, err := registry.GetAll(ctx)
	require.NoError(t, err)
	require.True(t, len(regData) == 0, "expected registry to be empty")

	testKey(ctx, t, fmt.Sprintf("%s-a", keyname), registry)
	testKey(ctx, t, fmt.Sprintf("%s-b", keyname), registry)
	testKey(ctx, t, fmt.Sprintf("%s-c", keyname), registry)
}

func TestNewRegistry(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	client, cleanUp := startContainerAndGetClient(ctx, t)
	defer cleanUp()

	t.Run("create empty global config", func(t *testing.T) {
		cmClient := client.CoreV1().ConfigMaps(namespace)

		_, err := cmClient.Get(ctx, createConfigName(globalConfigMapName), metav1.GetOptions{})
		assert.True(t, k8serrors.IsNotFound(err))

		globalReg, err := NewGlobalConfigRegistry(ctx, cmClient)
		require.NoError(t, err)

		assert.NotNil(t, globalReg)

		_, err = cmClient.Get(ctx, createConfigName(globalConfigMapName), metav1.GetOptions{})
		assert.NoError(t, err)
	})

	t.Run("create empty dogu config", func(t *testing.T) {
		doguName := "my-dogu"
		cmClient := client.CoreV1().ConfigMaps(namespace)

		_, err := cmClient.Get(ctx, createConfigName(doguName), metav1.GetOptions{})
		assert.True(t, k8serrors.IsNotFound(err))

		doguReg, err := NewDoguConfigRegistry(ctx, doguName, cmClient)
		require.NoError(t, err)

		assert.NotNil(t, doguReg)

		_, err = cmClient.Get(ctx, createConfigName(doguName), metav1.GetOptions{})
		assert.NoError(t, err)
	})

	t.Run("create empty sensitive dogu config", func(t *testing.T) {
		doguName := "my-dogu"
		sClient := client.CoreV1().Secrets(namespace)

		_, err := sClient.Get(ctx, createConfigName(doguName), metav1.GetOptions{})
		assert.True(t, k8serrors.IsNotFound(err))

		sensitiveReg, err := NewSensitiveDoguRegistry(ctx, doguName, sClient)
		require.NoError(t, err)

		assert.NotNil(t, sensitiveReg)

		_, err = sClient.Get(ctx, createConfigName(doguName), metav1.GetOptions{})
		assert.NoError(t, err)
	})

	t.Run("dont override existing registries", func(t *testing.T) {
		doguName := "my-dogu"

		const key, value = "config.yaml", "noValidYamlValue"

		sClient := client.CoreV1().Secrets(namespace)
		cmClient := client.CoreV1().ConfigMaps(namespace)

		// delete existing config maps
		err := cmClient.Delete(ctx, createConfigName(globalConfigMapName), metav1.DeleteOptions{})
		require.NoError(t, err)

		err = cmClient.Delete(ctx, createConfigName(doguName), metav1.DeleteOptions{})
		require.NoError(t, err)

		err = sClient.Delete(ctx, createConfigName(doguName), metav1.DeleteOptions{})
		require.NoError(t, err)

		// create global config
		exGlobalCfg, err := cmClient.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: createConfigName(globalConfigMapName)},
			Data:       map[string]string{key: value},
		}, metav1.CreateOptions{})
		require.NoError(t, err)

		// create dogu config
		exDoguCfg, err := cmClient.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: createConfigName(doguName)},
			Data:       map[string]string{key: value},
		}, metav1.CreateOptions{})
		require.NoError(t, err)

		// create sensitive dogu config
		exSenstiveCfg, err := sClient.Create(ctx, &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: createConfigName(doguName)},
			Data:       map[string][]byte{key: []byte(value)},
		}, metav1.CreateOptions{})
		require.NoError(t, err)

		// create registries
		_, err = NewGlobalConfigRegistry(ctx, cmClient)
		assert.NoError(t, err)

		_, err = NewDoguConfigRegistry(ctx, doguName, cmClient)
		assert.NoError(t, err)

		_, err = NewSensitiveDoguRegistry(ctx, doguName, sClient)
		assert.NoError(t, err)

		// evaluate entries in config maps and secret still exists
		globalCfgMap, err := cmClient.Get(ctx, createConfigName(globalConfigMapName), metav1.GetOptions{})
		assert.NoError(t, err)

		doguCfgMap, err := cmClient.Get(ctx, createConfigName(doguName), metav1.GetOptions{})
		assert.NoError(t, err)

		sensitiveCfg, err := sClient.Get(ctx, createConfigName(doguName), metav1.GetOptions{})
		assert.NoError(t, err)

		assert.Equal(t, exGlobalCfg, globalCfgMap)
		assert.Equal(t, exDoguCfg, doguCfgMap)
		assert.Equal(t, exSenstiveCfg, sensitiveCfg)
	})

}

func TestRegistry(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	client, cleanUp := startContainerAndGetClient(ctx, t)
	defer cleanUp()

	cmClient := client.CoreV1().ConfigMaps(namespace)
	sClient := client.CoreV1().Secrets(namespace)

	globalConfigRegistry, err := NewGlobalConfigRegistry(ctx, cmClient)
	require.NoError(t, err)

	doguConfigRegistry, err := NewDoguConfigRegistry(ctx, "test", cmClient)
	require.NoError(t, err)

	doguSecretRegistry, err := NewSensitiveDoguRegistry(ctx, "test", sClient)
	require.NoError(t, err)

	key := "key"
	testAllFunctions(ctx, t, globalConfigRegistry, key)
	testAllFunctions(ctx, t, doguConfigRegistry, key)
	testAllFunctions(ctx, t, doguSecretRegistry, key)
}

func TestRegistryWatch(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	client, cleanUp := startContainerAndGetClient(ctx, t)
	defer cleanUp()

	cmClient := client.CoreV1().ConfigMaps(namespace)

	t.Run("Watch global config registry", func(t *testing.T) {
		globalWatchCtx, gCancel := context.WithTimeout(ctx, 3*time.Second)
		defer gCancel()

		gRegistry, err := NewGlobalConfigRegistry(globalWatchCtx, cmClient)
		require.NoError(t, err, "failed to create global config registry")

		err = gRegistry.DeleteAll(globalWatchCtx)
		require.NoError(t, err)

		watchKey := "key1/key2"
		oldValue := "value"
		newValue := "newValue"

		err = gRegistry.Set(globalWatchCtx, watchKey, oldValue)
		require.NoError(t, err, "could not set initial value for global config")

		watch, err := gRegistry.Watch(globalWatchCtx, watchKey, false)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		waitCh := make(chan struct{})

		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-globalWatchCtx.Done():
				t.Errorf("context reach end before receiving watch result")
			case change := <-watch.ResultChan:
				assert.NoError(t, change.Err)
				values, ok := change.ModifiedKeys[watchKey]
				if !ok {
					t.Errorf("expected key %s in modifications", watchKey)
					return
				}

				assert.Equal(t, oldValue, values.OldValue)
				assert.Equal(t, newValue, values.NewValue)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			lErr := gRegistry.Set(globalWatchCtx, watchKey, newValue)
			require.NoError(t, lErr, "could not set new value for watchKey")
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

	t.Run("Watch global config registry with change on other key", func(t *testing.T) {
		globalWatchCtx, gCancel := context.WithTimeout(ctx, 3*time.Second)
		defer gCancel()

		gRegistry, err := NewGlobalConfigRegistry(globalWatchCtx, cmClient)
		require.NoError(t, err, "failed to create global config registry")

		err = gRegistry.DeleteAll(globalWatchCtx)
		require.NoError(t, err)

		watchKey := "key1/key2"
		oldValue := "value"
		newValue := "newValue"

		err = gRegistry.Set(globalWatchCtx, watchKey, oldValue)
		require.NoError(t, err, "could not set initial value for global config")

		watch, err := gRegistry.Watch(globalWatchCtx, watchKey, false)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		waitCh := make(chan struct{})

		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-globalWatchCtx.Done():
			case change := <-watch.ResultChan:
				t.Errorf("Received unexpected watch result %v", change)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			lErr := gRegistry.Set(globalWatchCtx, "/new", newValue)
			require.NoError(t, lErr, "could not set new value for watchKey")
		}()

		go func() {
			wg.Wait()
			close(waitCh)
		}()

		select {
		case <-globalWatchCtx.Done():
			t.Log("Got no notification about updates of new key")
		case <-waitCh:
			t.Errorf("Expected timeout because no change has happen on watchKey")
		}
	})

	t.Run("Recursive watch global config registry", func(t *testing.T) {
		globalWatchCtx, gCancel := context.WithTimeout(ctx, 3*time.Second)
		defer gCancel()

		gRegistry, err := NewGlobalConfigRegistry(globalWatchCtx, cmClient)
		require.NoError(t, err, "failed to create global config registry")

		err = gRegistry.DeleteAll(globalWatchCtx)
		require.NoError(t, err)

		watchKey := "key1"
		oldValue := "value"
		newValue := "newValue"

		err = gRegistry.Set(globalWatchCtx, fmt.Sprintf("%s/key2", watchKey), oldValue)
		require.NoError(t, err, "could not set initial value for global config")

		watch, err := gRegistry.Watch(globalWatchCtx, watchKey, true)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		waitCh := make(chan struct{})

		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-globalWatchCtx.Done():
				t.Errorf("context reach end before receiving watch result")
			case change := <-watch.ResultChan:
				assert.NoError(t, change.Err)
				assert.Equal(t, 1, len(change.ModifiedKeys))

				for k, v := range change.ModifiedKeys {
					assert.True(t, strings.Contains(k, watchKey))
					assert.Equal(t, "", v.OldValue)
					assert.Equal(t, newValue, v.NewValue)
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			lErr := gRegistry.Set(globalWatchCtx, fmt.Sprintf("%s/key3", watchKey), newValue)
			require.NoError(t, lErr, "could not set new value for watchKey")
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
