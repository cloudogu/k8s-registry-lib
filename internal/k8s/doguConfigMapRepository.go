package k8s

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
)

type DoguConfigMapRepo struct {
	configMapRepo
}

func CreateDoguConfigRepo(client ConfigMapClient) DoguConfigMapRepo {
	return DoguConfigMapRepo{
		configMapRepo: newConfigMapRepo(client, doguConfigType),
	}
}

func (dcmr DoguConfigMapRepo) GetDoguConfig(ctx context.Context, doguName string) (config.DoguConfig, error) {
	cfg, err := dcmr.getConfigByName(ctx, dcmr.createConfigName(doguName))
	if err != nil {
		return config.DoguConfig{}, err
	}

	return config.CreateDoguConfig(cfg), nil
}

func (dcmr DoguConfigMapRepo) WriteDoguConfigMap(ctx context.Context, cfg config.DoguConfig) error {
	return dcmr.writeConfig(ctx, cfg.Config)
}
