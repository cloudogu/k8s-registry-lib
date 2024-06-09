package k8s

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
)

const globalName = "global"

type GlobalConfigMapRepo struct {
	configMapRepo
}

func CreateGlobalConfigRepo(client ConfigMapClient) GlobalConfigMapRepo {
	return GlobalConfigMapRepo{
		configMapRepo: newConfigMapRepo(client, globalConfigType),
	}
}

func (gcmr GlobalConfigMapRepo) GetGlobalConfig(ctx context.Context) (config.GlobalConfig, error) {
	cfg, err := gcmr.getConfigByName(ctx, gcmr.createConfigName(globalName))
	if err != nil {
		return config.GlobalConfig{}, err
	}

	return config.CreateGlobalConfig(cfg), nil
}

func (gcmr GlobalConfigMapRepo) DeleteGlobalConfigMap(ctx context.Context) error {
	return gcmr.deleteConfigMap(ctx, globalName)
}

func (gcmr GlobalConfigMapRepo) WriteGlobalConfigMap(ctx context.Context, cfg config.GlobalConfig) error {
	return gcmr.writeConfig(ctx, cfg.Config)
}
