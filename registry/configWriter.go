package registry

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
)

type configWriter struct {
	repo configRepository
}

func (cw configWriter) Set(ctx context.Context, key, value string) error {
	doguConfig, err := cw.repo.get(ctx)
	if err != nil {
		if !errors.Is(err, ErrConfigNotFound) {
			return fmt.Errorf("could not read dogu config: %w", err)
		}

		//create new, empty doguConfig
		doguConfig = config.CreateConfig(make(config.Data))
	}

	doguConfig.Set(key, value)

	err = cw.repo.write(ctx, doguConfig)
	if err != nil {
		return fmt.Errorf("could not write dogu config after updating value: %w", err)
	}

	return nil
}

// Delete removes the configuration key and value
func (cw configWriter) Delete(ctx context.Context, key string) error {
	doguConfig, err := cw.repo.get(ctx)
	if err != nil {
		return fmt.Errorf("could not read dogu config: %w", err)
	}

	doguConfig.Delete(key)

	err = cw.repo.write(ctx, doguConfig)
	if err != nil {
		return fmt.Errorf("could not write dogu config after deleting key %s: %w", key, err)
	}

	return nil
}

// DeleteRecursive removes all configuration for the given key, including all configuration for sub-keys
func (cw configWriter) DeleteRecursive(ctx context.Context, key string) error {
	doguConfig, err := cw.repo.get(ctx)
	if err != nil {
		return fmt.Errorf("could not read dogu config: %w", err)
	}

	doguConfig.DeleteRecursive(key)

	err = cw.repo.write(ctx, doguConfig)
	if err != nil {
		return fmt.Errorf("could not write dogu config after recursively deleting key %s: %w", key, err)
	}

	return nil
}

func (cw configWriter) DeleteAll(ctx context.Context) error {
	doguConfig, err := cw.repo.get(ctx)
	if err != nil {
		return fmt.Errorf("could not read dogu config: %w", err)
	}

	doguConfig.DeleteAll()

	if lErr := cw.repo.write(ctx, doguConfig); lErr != nil {
		return fmt.Errorf("could not write dogu config after deleting all keys: %w", lErr)
	}

	return nil
}
