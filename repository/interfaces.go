package repository

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
)

type generalConfigRepository interface {
	get(context.Context, configName) (config.Config, error)
	delete(context.Context, configName) error
	create(context.Context, configName, config.SimpleDoguName, config.Config) (config.Config, error)
	update(context.Context, configName, config.SimpleDoguName, config.Config) (config.Config, error)
	saveOrMerge(context.Context, configName, config.Config) (config.Config, error)
	watch(ctx context.Context, name configName, filters ...config.WatchFilter) (<-chan configWatchResult, error)
}

type resourceVersionGetter interface {
	GetResourceVersion() string
}

type configClient interface {
	Get(ctx context.Context, name string) (clientData, error)
	Delete(ctx context.Context, name string) error
	Create(ctx context.Context, name string, doguName string, dataStr string) (resourceVersionGetter, error)
	Update(ctx context.Context, pCtx string, name string, doguName string, dataStr string) (resourceVersionGetter, error)
	UpdateClientData(ctx context.Context, update clientData) (resourceVersionGetter, error)
	Watch(ctx context.Context, name string) (<-chan clientWatchResult, error)
}
