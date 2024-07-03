package repository

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
)

const _SimpleGlobalConfigName = "global"

type GlobalConfigRepository struct {
	generalConfigRepository
}

func NewGlobalConfigRepository(client ConfigMapClient) *GlobalConfigRepository {
	cfgClient := createConfigMapClient(client, globalConfigType)
	cfgRepository := newConfigRepo(cfgClient)

	return &GlobalConfigRepository{
		generalConfigRepository: cfgRepository,
	}
}

func (gcr GlobalConfigRepository) Get(ctx context.Context) (config.GlobalConfig, error) {
	cfg, err := gcr.get(ctx, createConfigName(_SimpleGlobalConfigName))
	if err != nil {
		return config.GlobalConfig{}, fmt.Errorf("could not get global config: %w", err)
	}

	return config.GlobalConfig{
		Config: cfg,
	}, nil
}

func (gcr GlobalConfigRepository) Create(ctx context.Context, globalConfig config.GlobalConfig) (config.GlobalConfig, error) {
	cfg, err := gcr.create(ctx, createConfigName(_SimpleGlobalConfigName), "", globalConfig.Config)
	if err != nil {
		return config.GlobalConfig{}, fmt.Errorf("could not create global config: %w", err)
	}

	return config.GlobalConfig{
		Config: cfg,
	}, nil
}

func (gcr GlobalConfigRepository) Update(ctx context.Context, globalConfig config.GlobalConfig) (config.GlobalConfig, error) {
	cfg, err := gcr.update(ctx, createConfigName(_SimpleGlobalConfigName), "", globalConfig.Config)
	if err != nil {
		return config.GlobalConfig{}, fmt.Errorf("could not update global config: %w", err)
	}

	return config.GlobalConfig{
		Config: cfg,
	}, nil
}

func (gcr GlobalConfigRepository) SaveOrMerge(ctx context.Context, globalConfig config.GlobalConfig) (config.GlobalConfig, error) {
	cfg, err := gcr.saveOrMerge(ctx, createConfigName(_SimpleGlobalConfigName), globalConfig.Config)
	if err != nil {
		return config.GlobalConfig{}, fmt.Errorf("could not save and merge global config: %w", err)
	}

	return config.GlobalConfig{
		Config: cfg,
	}, nil
}

func (gcr GlobalConfigRepository) Delete(ctx context.Context) error {
	if err := gcr.delete(ctx, createConfigName(_SimpleGlobalConfigName)); err != nil {
		return fmt.Errorf("could not delete global config: %w", err)
	}

	return nil
}
