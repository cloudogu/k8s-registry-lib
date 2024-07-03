package repository

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
)

type SensitiveDoguConfigRepository struct {
	configRepository
}

func NewSensitiveDoguConfigRepository(client SecretClient) *SensitiveDoguConfigRepository {
	cfgClient := createSecretClient(client, sensitiveConfigType)
	cfgRepository := newConfigRepo(cfgClient)

	return &SensitiveDoguConfigRepository{
		configRepository: cfgRepository,
	}
}

func (scr SensitiveDoguConfigRepository) Get(ctx context.Context, name config.SimpleDoguName) (config.SensitiveDoguConfig, error) {
	cfg, err := scr.configRepository.get(ctx, createConfigName(name.String()))
	if err != nil {
		return config.SensitiveDoguConfig{}, fmt.Errorf("could not get sensitive config for dogu %s: %w", name.String(), err)
	}

	return config.SensitiveDoguConfig{
		DoguName: name,
		Config:   cfg,
	}, nil
}

func (scr SensitiveDoguConfigRepository) Create(ctx context.Context, sensitiveConfig config.SensitiveDoguConfig) (config.SensitiveDoguConfig, error) {
	doguName := sensitiveConfig.DoguName

	cfg, err := scr.configRepository.create(ctx, createConfigName(doguName.String()), doguName, sensitiveConfig.Config)
	if err != nil {
		return config.SensitiveDoguConfig{}, fmt.Errorf("could not create sensitive config for dogu %s: %w", doguName, err)
	}

	return config.SensitiveDoguConfig{
		DoguName: doguName,
		Config:   cfg,
	}, nil
}

func (scr SensitiveDoguConfigRepository) Update(ctx context.Context, sensitiveConfig config.SensitiveDoguConfig) (config.SensitiveDoguConfig, error) {
	doguName := sensitiveConfig.DoguName

	cfg, err := scr.configRepository.update(ctx, createConfigName(doguName.String()), doguName, sensitiveConfig.Config)
	if err != nil {
		return config.SensitiveDoguConfig{}, fmt.Errorf("could not update sensitive config for dogu %s: %w", doguName, err)
	}

	return config.SensitiveDoguConfig{
		DoguName: doguName,
		Config:   cfg,
	}, nil
}

func (scr SensitiveDoguConfigRepository) SaveOrMerge(ctx context.Context, sensitiveConfig config.SensitiveDoguConfig) (config.SensitiveDoguConfig, error) {
	cfg, err := scr.saveOrMerge(ctx, createConfigName(sensitiveConfig.DoguName.String()), sensitiveConfig.Config)
	if err != nil {
		return config.SensitiveDoguConfig{}, fmt.Errorf("could not save and merge sensitive config of dogu %s: %w", sensitiveConfig.DoguName, err)
	}

	return config.SensitiveDoguConfig{
		DoguName: sensitiveConfig.DoguName,
		Config:   cfg,
	}, nil
}

func (scr SensitiveDoguConfigRepository) Delete(ctx context.Context, name config.SimpleDoguName) error {
	if err := scr.configRepository.delete(ctx, createConfigName(name.String())); err != nil {
		return fmt.Errorf("could not delete sensitive config for dogu %s: %w", name, err)
	}

	return nil
}
