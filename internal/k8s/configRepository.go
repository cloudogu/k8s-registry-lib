package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
	"k8s.io/client-go/util/retry"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

var (
	ErrConfigNotFound = errors.New("could not find config")
)

type configType int

const (
	unknown configType = iota
	globalConfigType
	doguConfigType
	sensitiveConfigType
)

func (t configType) String() string {
	switch t {
	case globalConfigType:
		return "global-config"
	case doguConfigType:
		return "dogu-config"
	case sensitiveConfigType:
		return "sensitive-config"
	default:
		return "unknown"
	}
}

const (
	appLabelKey      = "app"
	appLabelValueCes = "ces"
	typeLabelKey     = "k8s.cloudogu.com/type"
)

const dataKeyName = "config.yaml"

type configData interface {
	get() map[string]string
}

type configClient interface {
	Get(ctx context.Context, name string) (configData, error)
	Delete(ctx context.Context, name string) error
	Create(ctx context.Context, name string, configData map[string]string, configType configType) error
	Update(ctx context.Context, configData configData) error
}

type configRepo struct {
	name       string
	client     configClient
	configType configType
	converter  config.Converter
}

func newConfigRepo(name string, client configClient, configType configType) configRepo {
	return configRepo{
		name:       name,
		client:     client,
		configType: configType,
		converter:  &config.YamlConverter{},
	}
}

func (cr configRepo) get(ctx context.Context) (config.Config, error) {
	if strings.TrimSpace(cr.name) == "" {
		return config.Config{}, errors.New("name is empty")
	}

	cd, err := cr.client.Get(ctx, cr.name)
	if err != nil {
		return config.Config{}, fmt.Errorf("unable to get config-map '%s' from cluster: %w", cr.name, err)
	}

	reader := strings.NewReader(cd.get()[dataKeyName])

	cfgData, err := cr.converter.Read(reader)
	if err != nil {
		return config.Config{}, fmt.Errorf("could not convert configmap data to config data: %w", err)
	}

	return config.CreateConfig(cfgData), nil
}

func (cr configRepo) delete(ctx context.Context) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := cr.client.Delete(ctx, cr.name); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("could not delete configmap in cluster: %w", err)
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

			return fmt.Errorf("unable to get current configmap with name %s: %w", cr.name, err)
		}

		return cr.updateConfig(ctx, cd, cfg)
	})
}

func (cr configRepo) createConfig(ctx context.Context, cfg config.Config) error {
	var buf bytes.Buffer

	if err := cr.converter.Write(&buf, cfg.Data); err != nil {
		return fmt.Errorf("unable to convert config data to configmap data: %w", err)
	}

	cd := map[string]string{
		dataKeyName: buf.String(),
	}

	if err := cr.client.Create(ctx, cr.name, cd, cr.configType); err != nil {
		return fmt.Errorf("could not create configmap in cluster: %w", err)
	}

	return nil
}

func (cr configRepo) updateConfig(ctx context.Context, cd configData, cfg config.Config) error {
	if len(cfg.ChangeHistory) == 0 {
		return nil
	}

	reader := strings.NewReader(cd.get()[dataKeyName])

	remoteConfigData, err := cr.converter.Read(reader)
	if err != nil {
		return fmt.Errorf("could not convert configmap data to config data: %w", err)
	}

	if reflect.DeepEqual(remoteConfigData, cfg.Data) {
		return nil
	}

	updatedRemoteConfigData, err := mergeConfigData(remoteConfigData, cfg)
	if err != nil {
		return fmt.Errorf("could not apply local changes to remote configmap: %w", err)
	}

	var buf bytes.Buffer

	if lErr := cr.converter.Write(&buf, updatedRemoteConfigData); lErr != nil {
		return fmt.Errorf("unable to convert config data to configmap data")
	}

	cd.get()[dataKeyName] = buf.String()

	if lErr := cr.client.Update(ctx, cd); lErr != nil {
		return fmt.Errorf("could not update configmap in cluster: %w", lErr)
	}

	return nil
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
