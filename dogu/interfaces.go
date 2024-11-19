package dogu

import (
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

type configMapClient interface {
	corev1client.ConfigMapInterface
}
