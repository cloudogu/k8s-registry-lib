package config

import (
	"fmt"
	"strings"
)

type Change struct {
	KeyPath string
	Deleted bool
}

type Data map[string]string

type Config struct {
	Name          string
	Data          Data
	ChangeHistory []Change
}

func CreateConfig(name string, data Data) Config {
	return Config{
		Name:          name,
		Data:          data,
		ChangeHistory: make([]Change, 0),
	}
}

func (c *Config) Set(key, value string) {
	c.Data[key] = value
	c.ChangeHistory = append(c.ChangeHistory, Change{KeyPath: key, Deleted: false})
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
	value, ok := c.Data[key]

	return value, ok
}

// GetAll returns a map of all key-value-pairs
func (c *Config) GetAll() Data {
	return c.Data
}

// Delete removes the configuration key and value
func (c *Config) Delete(key string) error {
	var keys []string

	for configKey := range c.Data {
		if strings.HasPrefix(configKey, key) {
			keys = append(keys, configKey)
		}
	}

	switch len(keys) {
	case 0:
		return nil
	case 1:
		delete(c.Data, key)
		c.ChangeHistory = append(c.ChangeHistory, Change{KeyPath: key, Deleted: true})

		return nil
	default:
		return fmt.Errorf("key %s does not point to single value", key)
	}
}

// DeleteRecursive removes all configuration for the given key, including all configuration for sub-keys
func (c *Config) DeleteRecursive(key string) {
	for configKey := range c.Data {
		if strings.HasPrefix(configKey, key) {
			delete(c.Data, configKey)
			c.ChangeHistory = append(c.ChangeHistory, Change{KeyPath: configKey, Deleted: true})
		}
	}
}

func (c *Config) RemoveAll() {
	c.DeleteRecursive("")
}

type GlobalConfig struct {
	Config
}

func CreateGlobalConfig(cfg Config) GlobalConfig {
	return GlobalConfig{
		Config: cfg,
	}
}

type DoguConfig struct {
	Config
}

func CreateDoguConfig(cfg Config) DoguConfig {
	return DoguConfig{
		Config: cfg,
	}
}
