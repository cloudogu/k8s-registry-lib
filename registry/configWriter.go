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
	cfg, err := cw.repo.get(ctx)
	if err != nil {
		if !errors.Is(err, ErrConfigNotFound) {
			return fmt.Errorf("could not read config: %w", err)
		}

		//create new, empty config
		cfg = config.CreateConfig(make(config.Data))
	}

	err = cfg.Set(key, value)
	if err != nil {
		return fmt.Errorf("could not set key %s with value %s: %w", key, value, err)
	}

	err = cw.repo.write(ctx, cfg)
	if err != nil {
		return fmt.Errorf("could not write config after updating value: %w", err)
	}

	return nil
}

// Delete removes the configuration key and value
func (cw configWriter) Delete(ctx context.Context, key string) error {
	cfg, err := cw.repo.get(ctx)
	if err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}

	cfg.Delete(key)

	err = cw.repo.write(ctx, cfg)
	if err != nil {
		return fmt.Errorf("could not write config after deleting key %s: %w", key, err)
	}

	return nil
}

// DeleteRecursive removes all configuration for the given key, including all configuration for sub-keys
func (cw configWriter) DeleteRecursive(ctx context.Context, key string) error {
	cfg, err := cw.repo.get(ctx)
	if err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}

	cfg.DeleteRecursive(key)

	err = cw.repo.write(ctx, cfg)
	if err != nil {
		return fmt.Errorf("could not write config after recursively deleting key %s: %w", key, err)
	}

	return nil
}

func (cw configWriter) DeleteAll(ctx context.Context) error {
	cfg, err := cw.repo.get(ctx)
	if err != nil {
		return fmt.Errorf("could not read dogu config: %w", err)
	}

	cfg.DeleteAll()

	if lErr := cw.repo.write(ctx, cfg); lErr != nil {
		return fmt.Errorf("could not write config after deleting all keys: %w", lErr)
	}

	return nil
}
