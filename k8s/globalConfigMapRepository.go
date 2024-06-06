package k8s

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GlobalConfigMapRepo struct {
	configMapRepo
}

func CreateGlobalConfigRepo(client ConfigMapClient) GlobalConfigMapRepo {
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

	dataStr := globalConfigMap.Data[dataKeyName]

	var yamlMap map[string]any
	err = yaml.Unmarshal([]byte(dataStr), &yamlMap)
	if err != nil {
		return GlobalConfig{}, fmt.Errorf("unable to parse yaml from global config map: %w", err)
	}

	var config map[string]string
	if lErr := flatMapToConfig(yamlMap, &config, ""); lErr != nil {
		return GlobalConfig{}, fmt.Errorf("cannot flat global config map: %w", err)
	}

	return CreateGlobalConfig(config), nil
}

func (gcmr GlobalConfigMapRepo) WriteGlobalConfigMap(ctx context.Context, cfg GlobalConfig) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		globalConfigMap, err := gcmr.client.Get(ctx, globalConfigName, metav1.GetOptions{})
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("unable to get current configmap for global config: %w", err)
		}

		if k8sErrs.IsNotFound(err) {
			return gcmr.createGlobalConfigMap(ctx, cfg)
		}

		return gcmr.updateGlobalConfigMap(ctx, globalConfigMap, cfg)
	})
}

func (gcmr GlobalConfigMapRepo) createGlobalConfigMap(ctx context.Context, cfg GlobalConfig) error {
	yamlData := configToMap(cfg.data, "")

	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		return fmt.Errorf("could not encode global config to yaml: %w", err)
	}

	yamlStr := string(yamlBytes)

	globalConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   globalConfigName,
			Labels: gcmr.labels,
		},
		Data: map[string]string{
			dataKeyName: yamlStr,
		},
	}

	_, lErr := gcmr.client.Create(ctx, globalConfigMap, metav1.CreateOptions{})
	if lErr != nil {
		return fmt.Errorf("could not create configmap for global config: %w", lErr)
	}

	return nil
}

func (gcmr GlobalConfigMapRepo) updateGlobalConfigMap(ctx context.Context, configMap *v1.ConfigMap, cfg GlobalConfig) error {
	if len(cfg.changeHistory) == 0 {
		return nil
	}

	var remoteConfigMap map[string]any
	err := yaml.Unmarshal([]byte(configMap.Data[dataKeyName]), &remoteConfigMap)
	if err != nil {
		return fmt.Errorf("could not parse remote global config to yaml: %w", err)
	}

	var remoteConfigData map[string]string
	err = flatMapToConfig(remoteConfigMap, &remoteConfigData, "")
	if err != nil {
		return fmt.Errorf("could not convert configmap to global config: %w", err)
	}

	if reflect.DeepEqual(remoteConfigData, cfg.data) {
		return nil
	}

	for _, c := range cfg.changeHistory {
		if c.Deleted {
			delete(remoteConfigData, c.KeyPath)
			continue
		}

		updatedValue, lErr := cfg.Get(c.KeyPath)
		if lErr != nil {
			return fmt.Errorf("unable get local global config value to write it to repository: %w", lErr)
		}

		remoteConfigData[c.KeyPath] = updatedValue
	}

	remoteConfigMap = configToMap(remoteConfigData, "")

	yamlBytes, err := yaml.Marshal(remoteConfigMap)
	if err != nil {
		return fmt.Errorf("could not encode global config to yaml: %w", err)
	}

	yamlStr := string(yamlBytes)

	configMap.Data[dataKeyName] = yamlStr

	_, err = gcmr.client.Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("could not update configmap for global config: %w", err)
	}

	return nil
}
