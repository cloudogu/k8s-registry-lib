package k8s

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const keySeparator = "/"
const dataKeyName = "config.yaml"
const globalConfigName = "globalConfig"

type GlobalConfig struct {
	YamlConfig
}

func CreateGlobalConfig() GlobalConfig {
	return GlobalConfig{
		YamlConfig: make(map[string]any),
	}
}

type GlobalConfigMapRepo struct {
	configMapRepo
}

func (gcmr GlobalConfigMapRepo) CreateGlobalConfigRepo(client ConfigMapClient) GlobalConfigMapRepo {
	return GlobalConfigMapRepo{
		configMapRepo: newConfigMapRepo(client, globalConfigType),
	}
}

func (gcmr GlobalConfigMapRepo) GetGlobalConfig(ctx context.Context) (cfg GlobalConfig, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("unable to get global config :%w", err)
		}
	}()

	globalConfigMap, err := gcmr.client.Get(ctx, globalConfigName, metav1.GetOptions{})
	if err != nil {
		if k8sErrs.IsNotFound(err) {
			return GlobalConfig{}, ErrConfigMapNotFound
		}

		return GlobalConfig{}, err
	}

	var yamlCfg YamlConfig
	dataStr := globalConfigMap.Data[dataKeyName]

	err = yaml.Unmarshal([]byte(dataStr), &yamlCfg)
	if err != nil {
		return GlobalConfig{}, fmt.Errorf("unable to parse string to global config: %w", err)
	}

	return GlobalConfig{YamlConfig: yamlCfg}, nil
}

func (gcmr GlobalConfigMapRepo) WriteGlobalConfigMap(ctx context.Context, cfg GlobalConfig) error {
	yamlBytes, err := yaml.Marshal(cfg.YamlConfig)
	if err != nil {
		return fmt.Errorf("could not encode global config to yaml: %w", err)
	}

	yamlStr := string(yamlBytes)

	globalConfigMap, err := gcmr.client.Get(ctx, globalConfigName, metav1.GetOptions{})
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("unable to get current configmap for global config: %w", err)
	}

	if k8sErrs.IsNotFound(err) {
		return gcmr.createGlobalConfigMap(ctx, yamlStr)
	}

	return gcmr.updateGlobalConfigMap(ctx, globalConfigMap, yamlStr)
}

func (gcmr GlobalConfigMapRepo) createGlobalConfigMap(ctx context.Context, yamlStr string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		globalConfigMap := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:   globalConfigName,
				Labels: gcmr.labels,
			},
			Data: map[string]string{
				dataKeyName: yamlStr,
			},
		}

		_, err := gcmr.client.Create(ctx, globalConfigMap, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("could not create configmap for global config: %w", err)
		}

		return nil
	})
}

func (gcmr GlobalConfigMapRepo) updateGlobalConfigMap(ctx context.Context, configMap *v1.ConfigMap, newConfigValue string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		configMap.Data[dataKeyName] = newConfigValue

		_, err := gcmr.client.Update(ctx, configMap, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("could not update configmap for global config: %w", err)
		}

		return nil
	})
}
