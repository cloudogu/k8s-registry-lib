package repository

import (
	"context"

	"github.com/cloudogu/k8s-registry-lib/config"

	informerCore "k8s.io/client-go/informers/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

type SecretClient interface {
	corev1client.SecretInterface
}

type SecretInformer interface {
	informerCore.SecretInformer
}

type ConfigMapClient interface {
	corev1client.ConfigMapInterface
}

type ConfigMapInformer interface {
	informerCore.ConfigMapInformer
}

type sharedInformer interface {
	cache.SharedIndexInformer
}

type generalConfigRepository interface {
	get(context.Context, configName) (config.Config, error)
	delete(context.Context, configName) error
	create(context.Context, configName, config.SimpleDoguName, config.Config) (config.Config, error)
	update(context.Context, configName, config.SimpleDoguName, config.Config) (config.Config, error)
	saveOrMerge(context.Context, configName, config.Config) (config.Config, error)
	watch(ctx context.Context, name configName, filters ...config.WatchFilter) (<-chan configWatchResult, error)
}
