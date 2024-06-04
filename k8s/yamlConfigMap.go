package k8s

import "strings"

type YamlConfig map[string]any

func (ycm *YamlConfig) Get(keyPath string) (string, bool) {
	keys := strings.Split(keyPath, keySeparator)
	var value any = ycm

	for _, key := range keys {
		if m, ok := value.(map[string]any); ok {
			value, ok = m[key]
			if !ok {
				return "", false
			}
		} else {
			return "", false
		}
	}

	return value.(string), true
}

func (ycm *YamlConfig) Set(keyPath string, value string) {
	keys := strings.Split(keyPath, keySeparator)
	lastKeyIndex := len(keys) - 1

	currentMap := *ycm
	for i, key := range keys {
		if i == lastKeyIndex {
			currentMap[key] = value
		} else {
			if nextMap, ok := currentMap[key].(map[string]any); ok {
				currentMap = nextMap
			} else {
				newMap := make(map[string]any)
				currentMap[key] = newMap
				currentMap = newMap
			}
		}
	}
}
