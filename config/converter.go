package config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"strings"
)

const keySeparator = "/"

type Converter interface {
	Read(reader io.Reader) (Data, error)
	Write(writer io.Writer, cfgData Data) error
}

func mapToConfig(sourceMap map[string]any, targetMapPtr *Data, parentPath string) error {
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
			err := mapToConfig(v, &targetMap, parentPath+sourceKey)
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

func configToMap(sourceMap Data, prefix string) map[string]any {
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

type YamlConverter struct {
}

func (yc *YamlConverter) Read(reader io.Reader) (Data, error) {
	if reader == nil {
		return nil, errors.New("reader is nil")
	}

	decoder := yaml.NewDecoder(reader)

	var yamlMap map[string]any
	if err := decoder.Decode(&yamlMap); err != nil {
		return nil, fmt.Errorf("unable to decode yaml from reader: %w", err)
	}

	var cfgData Data
	if err := mapToConfig(yamlMap, &cfgData, ""); err != nil {
		return nil, fmt.Errorf("cannot convert yaml Data to Config Data: %w", err)
	}

	return cfgData, nil
}

func (yc *YamlConverter) Write(writer io.Writer, cfgData Data) error {
	yamlMap := configToMap(cfgData, "")

	encoder := yaml.NewEncoder(writer)

	if err := encoder.Encode(yamlMap); err != nil {
		return fmt.Errorf("unable to encode Config Data as yaml to writer: %w", err)
	}

	return nil
}
