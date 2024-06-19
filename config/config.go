package config

import (
	"fmt"
	"maps"
	"strings"
)

type Change struct {
	KeyPath string
	Deleted bool
}

type Data map[string]string

type Config struct {
	Data          Data
	ChangeHistory []Change
}

func CreateConfig(data Data) Config {
	return Config{
		Data:          data,
		ChangeHistory: make([]Change, 0),
	}
}

func (c *Config) Set(key, value string) error {
	key = sanitizeKey(key)

	if key == "" || key == keySeparator {
		return fmt.Errorf("key is empty")
	}

	if strings.HasSuffix(key, keySeparator) {
		return fmt.Errorf("key %s must not be a dictionary", key)
	}

	subkey := key + keySeparator

	for configKey := range c.Data {
		if strings.HasPrefix(configKey, subkey) {
			return fmt.Errorf("key %s is alreaedy used as dictionary", configKey)
		}
	}

	if lErr := checkSubkeyHasValue(key, key, c.Data); lErr != nil {
		return fmt.Errorf("subkey from key %s already has a value: %w", key, lErr)
	}

	c.Data[key] = value
	c.ChangeHistory = append(c.ChangeHistory, Change{KeyPath: key, Deleted: false})

	return nil
}

func checkSubkeyHasValue(rootKey, key string, cfg Data) error {
	subkey, found := splitAtLastOccurrence(key)

	if _, ok := cfg[subkey]; ok && subkey != rootKey {
		return fmt.Errorf("key %s already has value set", subkey)
	}

	if found {
		return checkSubkeyHasValue(rootKey, subkey, cfg)
	}

	return nil
}

func splitAtLastOccurrence(s string) (string, bool) {
	// Find the last occurrence of the separator
	idx := strings.LastIndex(s, keySeparator)
	if idx == -1 {
		// If the separator is not found, return the original string and an empty string
		return s, false
	}
	// Split the string at the last occurrence of the separator
	return s[:idx], true
}

// Exists returns true if configuration key exists
func (c *Config) Exists(key string) bool {
	key = sanitizeKey(key)

	_, ok := c.Data[key]

	return ok
}

// Get returns the configuration value for the given key.
// Returns an error if no values exists for the given key.
func (c *Config) Get(key string) (string, error) {
	key = sanitizeKey(key)

	value, ok := c.Data[key]

	if !ok {
		return "", fmt.Errorf("value for %s does not exist", key)
	}

	return value, nil
}

// GetAll returns a map of all key-value-pairs
func (c *Config) GetAll() Data {
	return maps.Clone(c.Data)
}

// Delete removes the configuration key and value
func (c *Config) Delete(key string) {
	key = sanitizeKey(key)

	for configKey := range c.Data {
		if configKey == key {
			delete(c.Data, key)
			c.ChangeHistory = append(c.ChangeHistory, Change{KeyPath: key, Deleted: true})
		}
	}
}

// DeleteRecursive removes all configuration for the given key, including all configuration for sub-keys
func (c *Config) DeleteRecursive(key string) {
	key = sanitizeKey(key)

	c.Delete(key)

	//scan for subkeys
	if key != "" && !strings.HasSuffix(key, keySeparator) {
		key = key + keySeparator
	}

	for configKey := range c.Data {
		if strings.HasPrefix(configKey, key) {
			delete(c.Data, configKey)
			c.ChangeHistory = append(c.ChangeHistory, Change{KeyPath: configKey, Deleted: true})
		}
	}
}

func (c *Config) DeleteAll() {
	// delete recursive from root
	c.DeleteRecursive(keySeparator)
}

func sanitizeKey(key string) string {
	if strings.HasPrefix(key, keySeparator) {
		return key[1:]
	}

	return key
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
