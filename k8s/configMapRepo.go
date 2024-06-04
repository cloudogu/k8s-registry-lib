package k8s

import (
	"errors"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	ErrConfigMapNotFound = errors.New("could not find config map")
)

type configMapType int

const (
	unknown configMapType = iota
	doguRegistryType
	globalConfigType
	doguConfigType
)

func (t configMapType) String() string {
	switch t {
	case doguRegistryType:
		return "local-dogu-registry"
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

type configMapRepo struct {
	client ConfigMapClient
	labels labels.Set
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
