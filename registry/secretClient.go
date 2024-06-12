package registry

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretClient interface {
	corev1client.SecretInterface
}

type secretClient struct {
	client SecretClient
}

type configSecretData struct {
	s *v1.Secret
}

func (c *configSecretData) get() map[string]string {
	data := make(map[string]string)
	for k, v := range c.s.Data {
		data[k] = string(v)
	}

	return data
}

func (sc *secretClient) Get(ctx context.Context, name string) (configData, error) {
	secret, err := sc.client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8sErrs.IsNotFound(err) {
			return nil, ErrConfigNotFound
		}

		return nil, fmt.Errorf("unable to get secret from cluster: %w", err)
	}

	return &configSecretData{secret}, nil
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
		return fmt.Errorf("could not create secret in cluster: %w", err)
	}

	return nil
}

func (sc *secretClient) Update(ctx context.Context, cd configData) error {
	sd, ok := cd.(*configSecretData)
	if !ok {
		return fmt.Errorf("configData could not cast as secret")
	}

	if _, lErr := sc.client.Update(ctx, sd.s, metav1.UpdateOptions{}); lErr != nil {
		return fmt.Errorf("could not update secret in cluster: %w", lErr)
	}

	return nil
}
