package registry

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
	"golang.org/x/exp/maps"
	"strings"
)

type ConfigWatcher struct {
	repo configRepository
}

type ConfigModification struct {
	OldValue string
	NewValue string
}

type WatchResult struct {
	ModifiedKeys map[string]ConfigModification
	Err          error
}

// Watch watches for changes of the provided config-key and sends the event through the channel
func (cw ConfigWatcher) Watch(ctx context.Context, key string, recursive bool) (chan WatchResult, error) {
	confWatch, err := cw.repo.watch(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not watch config: %w", err)
	}

	lastConfig := confWatch.InitialConfig

	resultChan := make(chan WatchResult)

	go func() {
		for result := range confWatch.ResultChan {
			if result.err != nil {
				resultChan <- WatchResult{nil, fmt.Errorf("error watching config for key %s: %w", result.err, key)}
				continue
			}

			modifiedConfig := result.config

			modifications := compareConfigs(lastConfig, modifiedConfig, key, recursive)
			if len(modifications) > 0 {
				resultChan <- WatchResult{modifications, nil}
			}

			lastConfig = modifiedConfig
		}

		// watch-channel was closed
		close(resultChan)
		//FIXME what todo here???
	}()

	return resultChan, nil
}

func compareConfigs(oldConfig config.Config, newConfig config.Config, configKey string, recursive bool) map[string]ConfigModification {
	modifications := make(map[string]ConfigModification)

	if !recursive {
		mod, ok := compareConfigForSingleKey(oldConfig, newConfig, configKey)
		if ok {
			modifications[configKey] = mod
		}

		return modifications
	}

	combinedConfigMap := make(map[string]string)
	maps.Copy(combinedConfigMap, oldConfig.GetAll())
	maps.Copy(combinedConfigMap, newConfig.GetAll())
	allConfigKeys := maps.Keys(combinedConfigMap)

	for _, key := range allConfigKeys {
		if !strings.HasPrefix(key, configKey) {
			continue
		}

		mod, ok := compareConfigForSingleKey(oldConfig, newConfig, configKey)
		if ok {
			modifications[configKey] = mod
		}
	}

	return modifications
}

func compareConfigForSingleKey(oldConfig config.Config, newConfig config.Config, configKey string) (ConfigModification, bool) {
	oldValue, err := oldConfig.Get(configKey)
	if err != nil {
		oldValue = ""
	}

	newValue, err := newConfig.Get(configKey)
	if err != nil {
		newValue = ""
	}

	if oldValue != newValue {
		return ConfigModification{oldValue, newValue}, true
	}

	return ConfigModification{oldValue, newValue}, false
}
