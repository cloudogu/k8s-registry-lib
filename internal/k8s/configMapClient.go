package k8s

import (
	"context"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	errConfigMapNotFound = errors.New("could not find config-map")
)

type ConfigMapClient interface {
	corev1client.ConfigMapInterface
}

type configMapClient struct {
	client ConfigMapClient
}

type configMapData struct {
	cm *v1.ConfigMap
}

func (c *configMapData) get() map[string]string {
	return c.cm.Data
}

func (cmc *configMapClient) Get(ctx context.Context, name string) (configData, error) {
	cm, err := cmc.client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8sErrs.IsNotFound(err) {
			return nil, errConfigMapNotFound
		}

		return nil, fmt.Errorf("unable to get config-map from cluster: %w", err)
	}

	return &configMapData{cm}, nil
}

func (cmc *configMapClient) Delete(ctx context.Context, name string) error {
	if err := cmc.client.Delete(ctx, name, metav1.DeleteOptions{}); client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("could not delete config-map in cluster: %w", err)
	}

	return nil
}

func (cmc *configMapClient) Create(ctx context.Context, name string, configData map[string]string, configType configType) error {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: labels.Set{
				appLabelKey:  appLabelValueCes,
				typeLabelKey: configType.String(),
			},
		},
		Data: configData,
	}

	if _, err := cmc.client.Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("could not create configmap in cluster: %w", err)
	}

	return nil
}

func (cmc *configMapClient) Update(ctx context.Context, conf configData) error {
	co, ok := conf.(*configMapData)
	if !ok {
		return fmt.Errorf("configData could not cast as configMap")
	}

	if _, lErr := cmc.client.Update(ctx, co.cm, metav1.UpdateOptions{}); lErr != nil {
		return fmt.Errorf("could not update configmap in cluster: %w", lErr)
	}

	return nil
}
