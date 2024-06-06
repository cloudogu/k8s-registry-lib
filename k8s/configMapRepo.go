package k8s

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
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

const keySeparator = "/"
const dataKeyName = "config.yaml"

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

//TODO: Write general functions for yaml config repo so global and dogu repo can use it.

func flatMapToConfig(sourceMap map[string]interface{}, targetMapPtr *map[string]string, parentPath string) error {
	if targetMapPtr == nil {
		*targetMapPtr = make(map[string]string)
	}

	targetMap := *targetMapPtr

	if parentPath != "" {
		parentPath += keySeparator
	}
	for sourceKey, sourceValue := range sourceMap {
		switch v := sourceValue.(type) {
		case map[string]interface{}:
			err := flatMapToConfig(v, &targetMap, parentPath+sourceKey)
			if err != nil {
				return err
			}
		default:
			stringValue, ok := sourceValue.(string)
			if !ok {
				return fmt.Errorf("could not convert %v to string", sourceValue)
			}

			targetMap[parentPath+sourceKey] = stringValue
		}
	}

	return nil
}

func configToMap(sourceMap map[string]string, prefix string) map[string]any {
	interiorProps := make(map[string]bool)
	targetMap := make(map[string]any)
	for key := range sourceMap {
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			continue
		}

		key = strings.TrimPrefix(key, prefix)
		if strings.Contains(key, keySeparator) {
			interiorNode := strings.SplitN(key, keySeparator, 2)[0]
			interiorProps[interiorNode] = true
		} else {
			targetMap[key] = sourceMap[prefix+key]
		}
	}

	for key := range interiorProps {
		targetMap[key] = configToMap(sourceMap, prefix+key+keySeparator)
	}

	return targetMap
}
