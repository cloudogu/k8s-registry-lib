package registry

import (
	"context"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

type ConfigMapClient interface {
	corev1client.ConfigMapInterface
}

type configMapClient struct {
	client ConfigMapClient
	labels labels.Set
}

func createConfigMapClient(c ConfigMapClient, t configType) configMapClient {
	return configMapClient{
		client: c,
		labels: labels.Set{
			appLabelKey:  appLabelValueCes,
			typeLabelKey: t.String(),
		},
	}
}

func (cmc configMapClient) Get(ctx context.Context, name string) (clientData, error) {
	cm, err := cmc.client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8sErrs.IsNotFound(err) {
			return clientData{}, ErrConfigNotFound
		}

		return clientData{}, fmt.Errorf("unable to get config-map from cluster: %w", err)
	}

	dataStr, ok := cm.Data[dataKeyName]
	if !ok {
		return clientData{}, fmt.Errorf("could not find data for key %s", dataKeyName)
	}

	return clientData{
		dataStr: dataStr,
		rawData: cm,
	}, nil
}

func (cmc configMapClient) Delete(ctx context.Context, name string) error {
	if err := cmc.client.Delete(ctx, name, metav1.DeleteOptions{}); client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("could not delete config-map in cluster: %w", err)
	}

	return nil
}

func (cmc configMapClient) Create(ctx context.Context, name string, dataStr string) error {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: cmc.labels,
		},
		Data: map[string]string{
			dataKeyName: dataStr,
		},
	}

	if _, err := cmc.client.Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("could not create configmap in cluster: %w", err)
	}

	return nil
}

func (cmc configMapClient) Update(ctx context.Context, update clientData) error {
	cm, ok := update.rawData.(*v1.ConfigMap)
	if !ok {
		return fmt.Errorf("configData could not cast as configMap")
	}

	cm.Data[dataKeyName] = update.dataStr

	if _, err := cmc.client.Update(ctx, cm, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("could not update configmap in cluster: %w", err)
	}

	return nil
}

func (cmc configMapClient) Watch(ctx context.Context, name string) (*clientWatch, error) {
	fieldSelector := fmt.Sprintf("metadata.name=%s", name)
	cmWatch, err := cmc.client.Watch(ctx, metav1.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		return nil, fmt.Errorf("could not watch configmap '%s' in cluster: %w", name, err)
	}

	resultChan := make(chan clientWatchResult)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("[configmapClient] context was canceled")
				cmWatch.Stop()
				close(resultChan)
				return
			case result, ok := <-cmWatch.ResultChan():
				if !ok {
					fmt.Println("[configmapClient] resultChan was closed")
					// channel was closed
					cmWatch.Stop()
					close(resultChan)
					return
				}

				if result.Type == watch.Error {
					resultChan <- clientWatchResult{"", fmt.Errorf("error result in watch of configmap '%s'", name)}
					continue
				}

				cm, ok := result.Object.(*v1.ConfigMap)
				if !ok {
					resultChan <- clientWatchResult{"", fmt.Errorf("could not assert type of of configmap in watch")}
					continue
				}

				dataBytes, ok := cm.Data[dataKeyName]
				if !ok {
					resultChan <- clientWatchResult{"", fmt.Errorf("could not find data for key %s in configmap %s", dataKeyName, name)}
					continue
				}

				resultChan <- clientWatchResult{dataBytes, nil}

			}
		}
	}()

	return &clientWatch{resultChan}, nil
}

type SecretClient interface {
	corev1client.SecretInterface
}

type secretClient struct {
	client SecretClient
	labels labels.Set
}

func createSecretClient(c SecretClient, t configType) secretClient {
	return secretClient{
		client: c,
		labels: labels.Set{
			appLabelKey:  appLabelValueCes,
			typeLabelKey: t.String(),
		},
	}
}

func (sc secretClient) Get(ctx context.Context, name string) (clientData, error) {
	secret, err := sc.client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8sErrs.IsNotFound(err) {
			return clientData{}, ErrConfigNotFound
		}

		return clientData{}, fmt.Errorf("unable to get secret from cluster: %w", err)
	}

	dataBytes, ok := secret.Data[dataKeyName]
	if !ok {
		return clientData{}, fmt.Errorf("could not find data for key %s", dataKeyName)
	}

	return clientData{
		dataStr: string(dataBytes),
		rawData: secret,
	}, nil
}

func (sc secretClient) Delete(ctx context.Context, name string) error {
	if err := sc.client.Delete(ctx, name, metav1.DeleteOptions{}); client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("could not delete secret in cluster: %w", err)
	}

	return nil
}

func (sc secretClient) Create(ctx context.Context, name string, dataStr string) error {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: sc.labels,
		},
		StringData: map[string]string{
			dataKeyName: dataStr,
		},
	}

	if _, err := sc.client.Create(ctx, secret, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("could not create secret in cluster: %w", err)
	}

	return nil
}

func (sc secretClient) Update(ctx context.Context, update clientData) error {
	secret, ok := update.rawData.(*v1.Secret)
	if !ok {
		return fmt.Errorf("configData could not cast as secret")
	}

	secret.StringData = map[string]string{
		dataKeyName: update.dataStr,
	}

	if _, err := sc.client.Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("could not update secret in cluster: %w", err)
	}

	return nil
}

func (sc secretClient) Watch(ctx context.Context, name string) (*clientWatch, error) {
	fieldSelector := fmt.Sprintf("name=%s", name)
	secretWatch, err := sc.client.Watch(ctx, metav1.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		return nil, fmt.Errorf("could not watch secret '%s' in cluster: %w", name, err)
	}

	resultChan := make(chan clientWatchResult)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("[secretClient] context was canceled")

				secretWatch.Stop()
				close(resultChan)
				return
			case result, ok := <-secretWatch.ResultChan():
				if !ok {
					fmt.Println("[secretClient] resultChan was closed")

					// channel was closed
					secretWatch.Stop()
					close(resultChan)
					return
				}

				if result.Type == watch.Error {
					resultChan <- clientWatchResult{"", fmt.Errorf("error result in watch of secret '%s'", name)}
					continue
				}

				secret, ok := result.Object.(*v1.Secret)
				if !ok {
					resultChan <- clientWatchResult{"", fmt.Errorf("could not assert type of of secret in watch")}
					continue
				}

				dataBytes, ok := secret.Data[dataKeyName]
				if !ok {
					resultChan <- clientWatchResult{"", fmt.Errorf("could not find data for key %s in secret %s", dataKeyName, name)}
					continue
				}

				resultChan <- clientWatchResult{string(dataBytes), nil}
			}
		}
	}()

	return &clientWatch{resultChan}, nil
}
