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
	Read(reader io.Reader) (Entries, error)
	Write(writer io.Writer, cfgData Entries) error
}

func mapToConfig(sourceMap map[string]any, targetMapPtr *Entries, parentPath string) error {
	if *targetMapPtr == nil {
		*targetMapPtr = make(map[Key]Value)
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
		case string:
			targetMap[Key(parentPath+sourceKey)] = Value(v)
		default:
			return fmt.Errorf("could not convert %v to value (string)", sourceValue)
		}
	}

	return nil
}

func configToMap(sourceMap Entries, prefix string) map[string]any {
	interiorProps := make(map[string]bool)
	targetMap := make(map[string]any)
	for key := range sourceMap {
		if prefix != "" && !strings.HasPrefix(key.String(), prefix) {
			continue
		}

		key = Key(strings.TrimPrefix(key.String(), prefix))
		if strings.Contains(key.String(), keySeparator) {
			interiorNode := strings.SplitN(key.String(), keySeparator, 2)[0]
			interiorProps[interiorNode] = true
		} else {
			targetMap[key.String()] = sourceMap[Key(prefix)+key].String()
		}
	}

	for key := range interiorProps {
		targetMap[key] = configToMap(sourceMap, prefix+key+keySeparator)
	}

	return targetMap
}

type YamlConverter struct {
}

func (yc *YamlConverter) Read(reader io.Reader) (Entries, error) {
	if reader == nil {
		return nil, errors.New("reader is nil")
	}

	decoder := yaml.NewDecoder(reader)

	var yamlMap map[string]any
	if err := decoder.Decode(&yamlMap); err != nil {
		return nil, fmt.Errorf("unable to decode yaml from reader: %w", err)
	}

	return MapToEntries(yamlMap)
}

func MapToEntries(inputMap map[string]any) (Entries, error) {
	var cfgData Entries
	if err := mapToConfig(inputMap, &cfgData, ""); err != nil {
		return nil, fmt.Errorf("cannot convert map Entries to Config Entries: %w", err)
	}

	return cfgData, nil
}

func (yc *YamlConverter) Write(writer io.Writer, cfgData Entries) error {
	yamlMap := configToMap(cfgData, "")

	encoder := yaml.NewEncoder(writer)

	if err := encoder.Encode(yamlMap); err != nil {
		return fmt.Errorf("unable to encode Config Entries as yaml to writer: %w", err)
	}

	return nil
}
