package k8s

import corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

type ConfigMapClient interface {
	corev1client.ConfigMapInterface
}
