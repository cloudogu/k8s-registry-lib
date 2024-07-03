package repository

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
)

type DoguConfigRepository struct {
	configRepository
}

func NewDoguConfigRepository(client ConfigMapClient) *DoguConfigRepository {
	cfgClient := createConfigMapClient(client, doguConfigType)
	cfgRepository := newConfigRepo(cfgClient)

	return &DoguConfigRepository{
		configRepository: cfgRepository,
	}
}

func (dcr DoguConfigRepository) Get(ctx context.Context, name config.SimpleDoguName) (config.DoguConfig, error) {
	cfg, err := dcr.configRepository.get(ctx, createConfigName(name.String()))
	if err != nil {
		return config.DoguConfig{}, fmt.Errorf("could not get config for dogu %s: %w", name.String(), err)
	}

	return config.DoguConfig{
		DoguName: name,
		Config:   cfg,
	}, nil
}

func (dcr DoguConfigRepository) Create(ctx context.Context, doguConfig config.DoguConfig) (config.DoguConfig, error) {
	doguName := doguConfig.DoguName

	cfg, err := dcr.configRepository.create(ctx, createConfigName(doguName.String()), doguName, doguConfig.Config)
	if err != nil {
		return config.DoguConfig{}, fmt.Errorf("could not create config for dogu %s: %w", doguName, err)
	}

	return config.DoguConfig{
		DoguName: doguName,
		Config:   cfg,
	}, nil
}

func (dcr DoguConfigRepository) Update(ctx context.Context, doguConfig config.DoguConfig) (config.DoguConfig, error) {
	doguName := doguConfig.DoguName

	cfg, err := dcr.configRepository.update(ctx, createConfigName(doguName.String()), doguName, doguConfig.Config)
	if err != nil {
		return config.DoguConfig{}, fmt.Errorf("could not update config for dogu %s: %w", doguName, err)
	}

	return config.DoguConfig{
		DoguName: doguName,
		Config:   cfg,
	}, nil
}

func (dcr DoguConfigRepository) SaveOrMerge(ctx context.Context, doguConfig config.DoguConfig) (config.DoguConfig, error) {
	cfg, err := dcr.saveOrMerge(ctx, createConfigName(doguConfig.DoguName.String()), doguConfig.Config)
	if err != nil {
		return config.DoguConfig{}, fmt.Errorf("could not save and merge config of dogu %s: %w", doguConfig.DoguName, err)
	}

	return config.DoguConfig{
		DoguName: doguConfig.DoguName,
		Config:   cfg,
	}, nil
}

func (dcr DoguConfigRepository) Delete(ctx context.Context, name config.SimpleDoguName) error {
	if err := dcr.configRepository.delete(ctx, createConfigName(name.String())); err != nil {
		return fmt.Errorf("could not delete config for dogu %s: %w", name, err)
	}

	return nil
}
