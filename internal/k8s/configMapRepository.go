package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
	v1 "k8s.io/api/core/v1"
	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/retry"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

var (
	ErrConfigMapNotFound = errors.New("could not find config map")
)

type configMapType int

const (
	unknown configMapType = iota
	globalConfigType
	doguConfigType
)

func (t configMapType) String() string {
	switch t {
	case globalConfigType:
		return "global-config"
	case doguConfigType:
		return "dogu-config"
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

type configMapRepo struct {
	client    ConfigMapClient
	labels    labels.Set
	converter config.Converter
}

func newConfigMapRepo(client ConfigMapClient, mapType configMapType) configMapRepo {
	return configMapRepo{
		client: client,
		labels: labels.Set{
			appLabelKey:  appLabelValueCes,
			typeLabelKey: mapType.String(),
		},
	}
}

func (cmr configMapRepo) getConfigByName(ctx context.Context, name string) (config.Config, error) {
	if strings.TrimSpace(name) == "" {
		return config.Config{}, errors.New("name is empty")
	}

	configMap, err := cmr.client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8sErrs.IsNotFound(err) {
			return config.Config{}, ErrConfigMapNotFound
		}

		return config.Config{}, fmt.Errorf("unable to get config map from cluster: %w", err)
	}

	reader := strings.NewReader(configMap.Data[dataKeyName])

	cfgData, err := cmr.converter.Read(reader)
	if err != nil {
		return config.Config{}, fmt.Errorf("could not convert configmap data to config data: %w", err)
	}

	return config.CreateConfig(name, cfgData), nil
}

func (cmr configMapRepo) deleteConfigMap(ctx context.Context, name string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := cmr.client.Delete(ctx, name, metav1.DeleteOptions{}); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("could not delete configmap in cluster: %w", err)
		}

		return nil
	})
}

func (cmr configMapRepo) writeConfig(ctx context.Context, cfg config.Config) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		configMap, err := cmr.client.Get(ctx, cfg.Name, metav1.GetOptions{})
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("unable to get current configmap with name %s: %w", cfg.Name, err)
		}

		if k8sErrs.IsNotFound(err) {
			return cmr.createConfigMap(ctx, cfg)
		}

		return cmr.updateConfigMap(ctx, configMap, cfg)
	})
}

func (cmr configMapRepo) createConfigMap(ctx context.Context, cfg config.Config) error {
	var buf bytes.Buffer

	if err := cmr.converter.Write(&buf, cfg.Data); err != nil {
		return fmt.Errorf("unable to convert config data to configmap data: %w", err)
	}

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   cfg.Name,
			Labels: cmr.labels,
		},
		Data: map[string]string{
			dataKeyName: buf.String(),
		},
	}

	if _, err := cmr.client.Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("could not create configmap in cluster: %w", err)
	}

	return nil
}

func (cmr configMapRepo) updateConfigMap(ctx context.Context, configMap *v1.ConfigMap, cfg config.Config) error {
	if len(cfg.ChangeHistory) == 0 {
		return nil
	}

	reader := strings.NewReader(configMap.Data[dataKeyName])

	remoteConfigData, err := cmr.converter.Read(reader)
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

	if lErr := cmr.converter.Write(&buf, updatedRemoteConfigData); lErr != nil {
		return fmt.Errorf("unable to convert config data to configmap data")
	}

	configMap.Data[dataKeyName] = buf.String()

	if _, lErr := cmr.client.Update(ctx, configMap, metav1.UpdateOptions{}); lErr != nil {
		return fmt.Errorf("could not update configmap in cluster: %w", lErr)
	}

	return nil
}

func (cmr configMapRepo) createConfigName(name string) string {
	return fmt.Sprintf("%s-config", name)
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
