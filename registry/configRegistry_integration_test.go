//go:build integration
// +build integration

package registry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

		// evaluate Data in config maps and secret still exists
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
