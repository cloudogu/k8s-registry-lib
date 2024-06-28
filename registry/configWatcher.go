package registry

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
	"golang.org/x/exp/maps"
	"strings"
)

type configWatcher struct {
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

type ConfigWatch struct {
	ResultChan     chan WatchResult
	cancelWatchCtx context.CancelFunc
}

func (w ConfigWatch) Stop() {
	w.cancelWatchCtx()
}

// Watch watches for changes of the provided config-key and sends the event through the channel
func (cw configWatcher) Watch(ctx context.Context, key string, recursive bool) (ConfigWatch, error) {
	watchCtx, cancelWatchCtx := context.WithCancel(ctx)

	confWatch, err := cw.repo.watch(watchCtx)
	if err != nil {
		cancelWatchCtx()
		return ConfigWatch{}, fmt.Errorf("could not watch config: %w", err)
	}

	lastConfig := confWatch.InitialConfig

	resultChan := make(chan WatchResult)

	go func() {
		for result := range confWatch.ResultChan {
			if result.err != nil {
				resultChan <- WatchResult{nil, fmt.Errorf("error watching config for key %s: %w", key, result.err)}
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
	}()

	return ConfigWatch{resultChan, cancelWatchCtx}, nil
}

func compareConfigs(oldConfig config.Config, newConfig config.Config, sConfigKey string, recursive bool) map[string]ConfigModification {
	modifications := make(map[string]ConfigModification)

	configKey := config.Key(sConfigKey)

	if !recursive {
		mod, ok := compareConfigForSingleKey(oldConfig, newConfig, configKey)
		if ok {
			modifications[configKey.String()] = mod
		}

		return modifications
	}

	combinedConfigMap := make(map[config.Key]config.Value)
	maps.Copy(combinedConfigMap, oldConfig.GetAll())
	maps.Copy(combinedConfigMap, newConfig.GetAll())
	allConfigKeys := maps.Keys(combinedConfigMap)

	for _, key := range allConfigKeys {
		if !strings.HasPrefix(key.String(), configKey.String()) {
			continue
		}

		mod, ok := compareConfigForSingleKey(oldConfig, newConfig, key)
		if ok {
			modifications[key.String()] = mod
		}
	}

	return modifications
}

func compareConfigForSingleKey(oldConfig config.Config, newConfig config.Config, configKey config.Key) (ConfigModification, bool) {
	oldValue, err := oldConfig.Get(configKey)
	if err != nil {
		oldValue = ""
	}

	newValue, err := newConfig.Get(configKey)
	if err != nil {
		newValue = ""
	}

	if oldValue != newValue {
		return ConfigModification{oldValue.String(), newValue.String()}, true
	}

	return ConfigModification{oldValue.String(), newValue.String()}, false
}

type GlobalWatcher struct {
	configWatcher
}

type DoguWatcher struct {
	configWatcher
}

type SensitiveDoguWatcher struct {
	configWatcher
}

func NewGlobalConfigWatcher(ctx context.Context, k8sClient ConfigMapClient) (*GlobalWatcher, error) {
	repo, _ := newConfigRepo(withGeneralName(globalConfigMapName), createConfigMapClient(k8sClient, globalConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Entries))); lErr != nil {
		return nil, fmt.Errorf("could not create initial global config for global watcher: %w", lErr)
	}

	return &GlobalWatcher{
		configWatcher{repo: repo},
	}, nil
}

func NewDoguConfigWatcher(ctx context.Context, doguName string, k8sClient ConfigMapClient) (*DoguWatcher, error) {
	repo, _ := newConfigRepo(withDoguName(doguName), createConfigMapClient(k8sClient, doguConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Entries))); lErr != nil {
		return nil, fmt.Errorf("could not create initial dogu config for watcher %s: %w", doguName, lErr)
	}

	return &DoguWatcher{
		configWatcher{repo: repo},
	}, nil
}

func NewSensitiveDoguWatcher(ctx context.Context, doguName string, sc SecretClient) (*SensitiveDoguWatcher, error) {
	repo, _ := newConfigRepo(withDoguName(doguName), createSecretClient(sc, sensitiveConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Entries))); lErr != nil {
		return nil, fmt.Errorf("could not create initial sensitive dogu config for watcher %s: %w", doguName, lErr)
	}

	return &SensitiveDoguWatcher{
		configWatcher{repo: repo},
	}, nil
}
