package k8s

import (
	"fmt"
	"strings"
)

const globalConfigName = "global-config"

type Change struct {
	KeyPath string
	Deleted bool
}

type ConfigData map[string]string

type Config struct {
	name          string
	data          ConfigData
	changeHistory []Change
}

func (c *Config) Set(key, value string) {
	c.data[key] = value
	c.changeHistory = append(c.changeHistory, Change{KeyPath: key, Deleted: false})
}

// Exists returns true if configuration key exists
func (c *Config) Exists(key string) bool {
	_, ok := c.GetOrFalse(key)

	return ok
}

// Get returns the configuration value for the given key.
// Returns an error if no values exists for the given key.
func (c *Config) Get(key string) (string, error) {
	value, ok := c.GetOrFalse(key)

	if !ok {
		return "", fmt.Errorf("value for %s does not exist", key)
	}

	return value, nil
}

// GetOrFalse returns false and an empty string when the configuration value does not exist.
// Otherwise, returns true and the configuration value, even when the configuration value is an empty string.
func (c *Config) GetOrFalse(key string) (string, bool) {
	value, ok := c.data[key]

	return value, ok
}

// GetAll returns a map of all key-value-pairs
func (c *Config) GetAll() ConfigData {
	return c.data
}

// Delete removes the configuration key and value
func (c *Config) Delete(key string) error {
	var keys []string

	for configKey := range c.data {
		if strings.HasPrefix(configKey, key) {
			keys = append(keys, configKey)
		}
	}

	switch len(keys) {
	case 0:
		return nil
	case 1:
		delete(c.data, key)
		c.changeHistory = append(c.changeHistory, Change{KeyPath: key, Deleted: true})

		return nil
	default:
		return fmt.Errorf("key %s does not point to single value", key)
	}
}

// DeleteRecursive removes all configuration for the given key, including all configuration for sub-keys
func (c *Config) DeleteRecursive(key string) {
	for configKey := range c.data {
		if strings.HasPrefix(configKey, key) {
			delete(c.data, configKey)
			c.changeHistory = append(c.changeHistory, Change{KeyPath: configKey, Deleted: true})
		}
	}
}

type GlobalConfig struct {
	Config
}

func CreateGlobalConfig(cfgData ConfigData) GlobalConfig {
	return GlobalConfig{
		Config: Config{
			name:          globalConfigName,
			data:          cfgData,
			changeHistory: make([]Change, 0),
		},
	}
}

type DoguConfig struct {
	Config
}

func CreateDoguConfig(doguname string, cfgData ConfigData) DoguConfig {
	return DoguConfig{
		Config: Config{
			name:          fmt.Sprintf("%s-config", doguname),
			data:          cfgData,
			changeHistory: make([]Change, 0),
		},
	}
}
