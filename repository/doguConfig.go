package repository

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
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

// Get retrieves the config for the given dogu.
// It can throw the following errors:
//   - NotFoundError if the dogu config was not found
//   - ConnectionError if any connection problem happens, e.g. a timeout
//   - GenericError if any other error happens
func (dcr DoguConfigRepository) Get(ctx context.Context, name config.SimpleDoguName) (config.DoguConfig, error) {
	cfg, err := dcr.get(ctx, createConfigName(name.String()))
	if err != nil {
		return config.DoguConfig{}, fmt.Errorf("could not get config for dogu %s: %w", name.String(), err)
	}

	return config.DoguConfig{
		DoguName: name,
		Config:   cfg,
	}, nil
}

// Create initially creates the underlying data structure for the dogu config.
// It can throw the following errors:
//   - AlreadyExistsError if the dogu config was already created
//   - ConnectionError if any connection problem happens, e.g. a timeout
//   - GenericError if any other error happens
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

// Update persists the dogu config as a whole.
// It can throw the following errors:
//   - ConflictError if there were concurrent updates to the config
//   - ConnectionError if any connection problem happens, e.g. a timeout
//   - GenericError if any other error happens
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

// SaveOrMerge persists the dogu config with a merge approach at conflicts. This means, that only keys will be overwritten,
// that got set explicitly. Therefore, the stored config could vary significantly and can be in an inconsistent state.
// You will not get any hint that a merge happened.
// Only use this function if you are absolutely sure, that you are allowed to take this risk.
// It can throw the following errors:
//   - ConnectionError if any connection problem happens, e.g. a timeout
//   - GenericError if any other error happens
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

// Delete removes the underlying data structure for the dogu config.
// It can throw the following errors:
//   - ConnectionError if any connection problem happens, e.g. a timeout
//   - GenericError if any other error happens
//
// If the config is not found, no error will happen as it would be deleted anyway (idempotent).
func (dcr DoguConfigRepository) Delete(ctx context.Context, name config.SimpleDoguName) error {
	if err := dcr.delete(ctx, createConfigName(name.String())); err != nil {
		return fmt.Errorf("could not delete config for dogu %s: %w", name, err)
	}

	return nil
}

// DoguConfigWatchResult can be used to create a diff of the config and react to possible changed keys.
type DoguConfigWatchResult struct {
	// PrevState is the state of the config before the current change
	PrevState config.DoguConfig
	// NewState is the state of the config after the current change
	NewState config.DoguConfig
	Err      error
}

// Watch returns a channel for DoguConfigWatchResult's which you can use to get informed about any changes to the dogu config.
// You can also optionally apply config.WatchFilter's to not get informed about uninteresting changes.
// The watcher automatically tries to reconnect if a connectionError happens. It does it in a way, that you don't
// miss any events. However, the events could be delayed or piled up, so updates in reaction to changes is
// bad practice and will lead to conflictErrors eventually.
// It can throw the following errors:
//   - NotFoundError if the dogu config was not found
//   - ConnectionError if any connection problem happens, e.g. a timeout, and the watcher could not reconnect automatically
//   - GenericError if any other error happens
//
// Please check for possible errors in the DoguConfigWatchResult too.
func (dcr DoguConfigRepository) Watch(ctx context.Context, dName config.SimpleDoguName, filters ...config.WatchFilter) (<-chan DoguConfigWatchResult, error) {
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
