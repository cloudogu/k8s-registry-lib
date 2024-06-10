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
	errSecretNotFound = errors.New("could not find secret")
)

type SecretClient interface {
	corev1client.SecretInterface
}

type secretClient struct {
	client SecretClient
}

func (sc *secretClient) Get(ctx context.Context, name string) (map[string]string, error) {
	secret, err := sc.client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8sErrs.IsNotFound(err) {
			return nil, errSecretNotFound
		}

		return nil, fmt.Errorf("unable to get secret from cluster: %w", err)
	}

	data := make(map[string]string)
	for k, v := range secret.Data {
		data[k] = string(v)
	}

	return data, nil
}

func (sc *secretClient) Delete(ctx context.Context, name string) error {
	if err := sc.client.Delete(ctx, name, metav1.DeleteOptions{}); client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("could not delete secret in cluster: %w", err)
	}

	return nil
}

func (sc *secretClient) Create(ctx context.Context, name string, configData map[string]string, configType configType) error {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: labels.Set{
				appLabelKey:  appLabelValueCes,
				typeLabelKey: configType.String(),
			},
		},
		StringData: configData,
	}

	if _, err := sc.client.Create(ctx, secret, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("could not create configmap in cluster: %w", err)
	}

	return nil
}

func (sc *secretClient) Update(ctx context.Context, name string, configData map[string]string) error {
	secret, err := sc.client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to update secret from cluster: %w", err)
	}

	secret.StringData = configData

	if _, lErr := sc.client.Update(ctx, secret, metav1.UpdateOptions{}); lErr != nil {
		return fmt.Errorf("could not update configmap in cluster: %w", lErr)
	}

	return nil
}
