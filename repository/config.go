package repository

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
	"reflect"
	"strings"
)

type configName string

func (c configName) String() string {
	return string(c)
}

func createConfigName(simpleName string) configName {
	return configName(strings.ToLower(fmt.Sprintf("%s-config", simpleName)))
}

type configRepository struct {
	client    configClient
	converter config.Converter
}

var _ generalConfigRepository = configRepository{}

func newConfigRepo(client configClient) configRepository {
	cr := configRepository{
		client:    client,
		converter: &config.YamlConverter{},
	}

	return cr
}

func (cr configRepository) get(ctx context.Context, name configName) (config.Config, error) {
	cd, err := cr.client.Get(ctx, name.String())
	if err != nil {
		return config.Config{}, fmt.Errorf("unable to get data '%s' from cluster: %w", name, err)
	}

	reader := strings.NewReader(cd.dataStr)

	cfgData, err := cr.converter.Read(reader)
	if err != nil {
		return config.Config{}, fmt.Errorf("could not convert client data to config data: %w", err)
	}

	cfg := config.CreateConfig(
		cfgData,
		config.WithPersistenceContext(getPersistentContext(cd.rawData)),
	)

	return cfg, nil
}

func (cr configRepository) delete(ctx context.Context, name configName) error {
	if err := cr.client.Delete(ctx, name.String()); err != nil {
		return fmt.Errorf("could not delete data '%s' in cluster: %w", name, err)
	}

	return nil
}

func (cr configRepository) create(ctx context.Context, name configName, doguName config.SimpleDoguName, cfg config.Config) (config.Config, error) {
	var buf bytes.Buffer

	if err := cr.converter.Write(&buf, cfg.GetAll()); err != nil {
		return config.Config{}, fmt.Errorf("unable to convert config data to data string: %w", err)
	}

	resource, err := cr.client.Create(ctx, name.String(), doguName.String(), buf.String())
	if err != nil {
		return config.Config{}, fmt.Errorf("could not create config in cluster: %w", err)
	}

	cfg.PersistenceContext = resource.GetResourceVersion()

	return cfg, nil
}

func (cr configRepository) update(ctx context.Context, name configName, doguName config.SimpleDoguName, cfg config.Config) (config.Config, error) {
	var buf bytes.Buffer

	if err := cr.converter.Write(&buf, cfg.GetAll()); err != nil {
		return config.Config{}, fmt.Errorf("unable to convert config data to data string: %w", err)
	}

	resource, err := cr.client.Update(ctx, getPersistentContext(cfg.PersistenceContext), name.String(), doguName.String(), buf.String())
	if err != nil {
		return config.Config{}, fmt.Errorf("could not update config in cluster: %w", err)
	}

	cfg.PersistenceContext = resource.GetResourceVersion()

	return cfg, nil
}

func (cr configRepository) saveOrMerge(ctx context.Context, name configName, cfg config.Config) (config.Config, error) {
	if len(cfg.GetChangeHistory()) == 0 {
		return cfg, nil
	}

	cd, err := cr.client.Get(ctx, name.String())
	if err != nil {
		return config.Config{}, fmt.Errorf("unable to get current data with name '%s' from cluster: %w", name, err)
	}

	reader := strings.NewReader(cd.dataStr)

	remoteConfigData, err := cr.converter.Read(reader)
	if err != nil {
		return config.Config{}, fmt.Errorf("could not convert old client data to config data: %w", err)
	}

	if reflect.DeepEqual(remoteConfigData, cfg.GetAll()) {
		return cfg, nil
	}

	updatedRemoteConfigData, err := mergeConfigData(remoteConfigData, cfg)
	if err != nil {
		return config.Config{}, fmt.Errorf("could not apply local changes to remote data: %w", err)
	}

	var buf bytes.Buffer

	if lErr := cr.converter.Write(&buf, updatedRemoteConfigData); lErr != nil {
		return config.Config{}, fmt.Errorf("unable to convert config data to data string: %w", lErr)
	}

	cd.dataStr = buf.String()

	updatedResource, err := cr.client.UpdateClientData(ctx, cd)
	if err != nil {
		return config.Config{}, fmt.Errorf("could not update data in cluster: %w", err)
	}

	updatedConfig := config.CreateConfig(
		updatedRemoteConfigData,
		config.WithPersistenceContext(getPersistentContext(updatedResource)),
	)

	return updatedConfig, nil
}

func mergeConfigData(remoteCfgData config.Entries, localCfg config.Config) (config.Entries, error) {
	for _, c := range localCfg.GetChangeHistory() {
		if c.Deleted {
			delete(remoteCfgData, c.KeyPath)
			continue
		}

		updatedValue, ok := localCfg.Get(c.KeyPath)
		if !ok {
			return nil, fmt.Errorf("unable get local config value for key %s to update remote config value", c.KeyPath)
		}

		remoteCfgData[c.KeyPath] = updatedValue
	}

	return remoteCfgData, nil
}

type configWatchResult struct {
	prevState config.Config
	newState  config.Config
	err       error
}

func (cr configRepository) watch(ctx context.Context, name configName, filters ...config.WatchFilter) (<-chan configWatchResult, error) {
	lastCfg, err := cr.get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("could not get config: %w", err)
	}

	clientResultChan, err := cr.client.Watch(ctx, name.String(), lastCfg.PersistenceContext)
	if err != nil {
		return nil, fmt.Errorf("could not start watch: %w", err)
	}

	resultChan := make(chan configWatchResult)

	go func() {
		defer close(resultChan)
		for clientResult := range clientResultChan {
			configResult := createConfigWatchResult(lastCfg, clientResult, cr.converter)

			if configResult.err != nil {
				resultChan <- configResult
				continue
			}

			// when no filter is set, notify about every change
			if len(filters) == 0 {
				resultChan <- configResult
				lastCfg = configResult.newState
				continue
			}

			// apply filters, notify if one of the filters matches
			for _, filter := range filters {
				if filter(configResult.prevState.Diff(configResult.newState)) {
					resultChan <- configResult
					lastCfg = configResult.newState

					break
				}
			}
		}
	}()

	return resultChan, nil
}

func createConfigWatchResult(lastCfg config.Config, result clientWatchResult, converter config.Converter) configWatchResult {
	if result.err != nil {
		return configWatchResult{
			prevState: config.Config{},
			newState:  config.Config{},
			err:       fmt.Errorf("client watch error: %w", result.err),
		}
	}

	reader := strings.NewReader(result.dataStr)

	cfgData, err := converter.Read(reader)
	if err != nil {
		return configWatchResult{
			prevState: config.Config{},
			newState:  config.Config{},
			err:       fmt.Errorf("could not convert client data to config data: %w", err),
		}
	}

	return configWatchResult{
		prevState: lastCfg,
		newState:  config.CreateConfig(cfgData, config.WithPersistenceContext(result.persistentContext)),
		err:       nil,
	}
}

func getPersistentContext(rawData any) string {
	switch r := rawData.(type) {
	case string:
		return r
	case resourceVersionGetter:
		return r.GetResourceVersion()
	default:
		return ""
	}
}
