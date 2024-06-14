package registry

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
	"k8s.io/client-go/util/retry"
	"reflect"
	"strings"
)

type clientData struct {
	dataStr string
	rawData any
}

type clientWatchResult struct {
	dataStr string
	err     error
}

type configWatch struct {
	InitialConfig config.Config
	ResultChan    <-chan configWatchResult
}

type configWatchResult struct {
	config config.Config
	err    error
}

type configClient interface {
	Get(ctx context.Context, name string) (clientData, error)
	Delete(ctx context.Context, name string) error
	Create(ctx context.Context, name string, dataStr string) error
	Update(ctx context.Context, update clientData) error
	Watch(ctx context.Context, name string) (<-chan clientWatchResult, error)
}

type configRepo struct {
	name      string
	client    configClient
	converter config.Converter
}

func newConfigRepo(name string, client configClient) (configRepo, error) {
	if strings.TrimSpace(name) == "" {
		return configRepo{}, errors.New("name is empty")
	}

	if client == nil {
		return configRepo{}, errors.New("client is nil")
	}

	return configRepo{
		name:      name,
		client:    client,
		converter: &config.YamlConverter{},
	}, nil
}

func (cr configRepo) get(ctx context.Context) (config.Config, error) {
	cd, err := cr.client.Get(ctx, cr.name)
	if err != nil {
		return config.Config{}, fmt.Errorf("unable to get data '%s' from cluster: %w", cr.name, err)
	}

	reader := strings.NewReader(cd.dataStr)

	cfgData, err := cr.converter.Read(reader)
	if err != nil {
		return config.Config{}, fmt.Errorf("could not convert client data to config data: %w", err)
	}

	return config.CreateConfig(cfgData), nil
}

func (cr configRepo) delete(ctx context.Context) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := cr.client.Delete(ctx, cr.name); err != nil {
			return fmt.Errorf("could not delete data '%s' in cluster: %w", cr.name, err)
		}

		return nil
	})
}

func (cr configRepo) write(ctx context.Context, cfg config.Config) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cd, err := cr.client.Get(ctx, cr.name)
		if err != nil {
			if errors.Is(err, ErrConfigNotFound) {
				return cr.createConfig(ctx, cfg)
			}

			return fmt.Errorf("unable to get current data with name '%s' from cluster: %w", cr.name, err)
		}

		return cr.updateConfig(ctx, cd, cfg)
	})
}

func (cr configRepo) createConfig(ctx context.Context, cfg config.Config) error {
	var buf bytes.Buffer

	if err := cr.converter.Write(&buf, cfg.Data); err != nil {
		return fmt.Errorf("unable to convert config data to data string: %w", err)
	}

	if err := cr.client.Create(ctx, cr.name, buf.String()); err != nil {
		return fmt.Errorf("could not create config in cluster: %w", err)
	}

	return nil
}

func (cr configRepo) updateConfig(ctx context.Context, cd clientData, cfg config.Config) error {
	if len(cfg.ChangeHistory) == 0 {
		return nil
	}

	reader := strings.NewReader(cd.dataStr)

	remoteConfigData, err := cr.converter.Read(reader)
	if err != nil {
		return fmt.Errorf("could not convert old client data to config data: %w", err)
	}

	if reflect.DeepEqual(remoteConfigData, cfg.Data) {
		return nil
	}

	updatedRemoteConfigData, err := mergeConfigData(remoteConfigData, cfg)
	if err != nil {
		return fmt.Errorf("could not apply local changes to remote data: %w", err)
	}

	var buf bytes.Buffer

	if lErr := cr.converter.Write(&buf, updatedRemoteConfigData); lErr != nil {
		return fmt.Errorf("unable to convert config data to data string")
	}

	cd.dataStr = buf.String()

	if lErr := cr.client.Update(ctx, cd); lErr != nil {
		return fmt.Errorf("could not update data in cluster: %w", lErr)
	}

	return nil
}

func (cr configRepo) watch(ctx context.Context) (*configWatch, error) {
	current, err := cr.get(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get config: %w", err)
	}

	clientResultChan, err := cr.client.Watch(ctx, cr.name)
	if err != nil {
		return nil, fmt.Errorf("could not start watch: %w", err)
	}

	resultChan := make(chan configWatchResult)

	go func() {
		for result := range clientResultChan {
			if result.err != nil {
				resultChan <- configWatchResult{config.Config{}, fmt.Errorf("error watching config: %w", result.err)}
				continue
			}

			reader := strings.NewReader(result.dataStr)
			cfgData, rErr := cr.converter.Read(reader)
			if rErr != nil {
				resultChan <- configWatchResult{config.Config{}, fmt.Errorf("could not convert client data to config data: %w", rErr)}
				continue
			}

			resultChan <- configWatchResult{config.CreateConfig(cfgData), nil}
		}

		close(resultChan)
	}()

	return &configWatch{
		InitialConfig: current,
		ResultChan:    resultChan,
	}, nil
}

func mergeConfigData(remoteCfgData config.Data, localCfg config.Config) (config.Data, error) {
	for _, c := range localCfg.ChangeHistory {
		if c.Deleted {
			delete(remoteCfgData, c.KeyPath)
			continue
		}

		updatedValue, lErr := localCfg.Get(c.KeyPath)
		if lErr != nil {
			return nil, fmt.Errorf("unable get local config value to update remote config value: %w", lErr)
		}

		remoteCfgData[c.KeyPath] = updatedValue
	}

	return remoteCfgData, nil
}
