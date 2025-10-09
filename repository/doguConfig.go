package repository

import (
	"context"
	"fmt"

	"github.com/cloudogu/ces-commons-lib/dogu"
	"github.com/cloudogu/k8s-registry-lib/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DoguConfigRepository struct {
	generalConfigRepository
}

func NewDoguConfigRepository(client ConfigMapClient) *DoguConfigRepository {
	cfgClient := createConfigMapClient(client, doguConfigType)
	cfgRepository := newConfigRepo(cfgClient)

	return &DoguConfigRepository{
		generalConfigRepository: cfgRepository,
	}
}

func NewSensitiveDoguConfigRepository(client SecretClient) *DoguConfigRepository {
	cfgClient := createSecretClient(client, sensitiveConfigType)
	cfgRepository := newConfigRepo(cfgClient)

	return &DoguConfigRepository{
		generalConfigRepository: cfgRepository,
	}
}

func (dcr DoguConfigRepository) Get(ctx context.Context, name dogu.SimpleName) (config.DoguConfig, error) {
	cfg, err := dcr.get(ctx, createConfigName(name.String()))
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

	cfg, err := dcr.create(ctx, createConfigName(doguName.String()), doguName, doguConfig.Config)
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

	cfg, err := dcr.update(ctx, createConfigName(doguName.String()), doguName, doguConfig.Config)
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

func (dcr DoguConfigRepository) Delete(ctx context.Context, name dogu.SimpleName) error {
	if err := dcr.delete(ctx, createConfigName(name.String())); err != nil {
		return fmt.Errorf("could not delete config for dogu %s: %w", name, err)
	}

	return nil
}

type DoguConfigWatchResult struct {
	PrevState config.DoguConfig
	NewState  config.DoguConfig
	Err       error
}

func (dcr DoguConfigRepository) Watch(ctx context.Context, dName dogu.SimpleName, filters ...config.WatchFilter) (<-chan DoguConfigWatchResult, error) {
	cfgWatch, err := dcr.watch(ctx, createConfigName(dName.String()), filters...)
	if err != nil {
		return nil, fmt.Errorf("unable to start watch for config from dogu %s: %w", dName, err)
	}

	watchChan := make(chan DoguConfigWatchResult)

	go func() {
		defer close(watchChan)
		for result := range cfgWatch {
			watchChan <- DoguConfigWatchResult{
				PrevState: config.DoguConfig{
					DoguName: dName,
					Config:   result.prevState,
				},
				NewState: config.DoguConfig{
					DoguName: dName,
					Config:   result.newState,
				},
				Err: result.err,
			}
		}
	}()

	return watchChan, nil
}

func (dcr DoguConfigRepository) SetOwnerReference(ctx context.Context, dName dogu.SimpleName, owners []metav1.OwnerReference) error {
	err := dcr.setOwnerReference(ctx, createConfigName(dName.String()), owners)
	if err != nil {
		return fmt.Errorf("failed to update owner reference for dogu config for dogu %s: %w", dName, err)
	}
	return nil
}
