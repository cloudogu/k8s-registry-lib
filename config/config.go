package config

import (
	"fmt"
	"golang.org/x/exp/slices"
	"maps"
	"strings"
)

type Change struct {
	KeyPath Key
	Deleted bool
}

type Key string

func (k Key) String() string {
	return string(k)
}

type Value string

func (v Value) String() string {
	return string(v)
}

type Entries map[Key]Value

type Config struct {
	entries       Entries
	changeHistory []Change
}

func CreateConfig(data Entries) Config {
	return Config{
		entries:       data,
		changeHistory: make([]Change, 0),
	}
}

func (c *Config) Set(k Key, v Value) error {
	k = sanitizeKey(k)

	if k == "" || k == keySeparator {
		return fmt.Errorf("key is empty")
	}

	if strings.HasSuffix(k.String(), keySeparator) {
		return fmt.Errorf("key %s must not be a dictionary", k)
	}

	subKey := k + keySeparator

	for configKey := range c.entries {
		if strings.HasPrefix(configKey.String(), subKey.String()) {
			return fmt.Errorf("key %s is alreaedy used as dictionary", configKey)
		}
	}

	if lErr := validateNoDictionaryHasValue(k, k, c.entries); lErr != nil {
		return fmt.Errorf("dictionary with key %s already has a value: %w", k, lErr)
	}

	c.entries[k] = v
	c.changeHistory = append(c.changeHistory, Change{KeyPath: k, Deleted: false})

	return nil
}

func validateNoDictionaryHasValue(rootKey, key Key, cfg Entries) error {
	subKey, found := splitAtLastOccurrence(key)

	if _, ok := cfg[subKey]; ok && subKey != rootKey {
		return fmt.Errorf("key %s already has Value set", subKey)
	}

	if found {
		return validateNoDictionaryHasValue(rootKey, subKey, cfg)
	}

	return nil
}

func splitAtLastOccurrence(s Key) (Key, bool) {
	// Find the last occurrence of the separator
	idx := strings.LastIndex(s.String(), keySeparator)
	if idx == -1 {
		// If the separator is not found, return the original string and an empty string
		return s, false
	}
	// Split the string at the last occurrence of the separator
	return s[:idx], true
}

// Exists returns true if configuration Key exists
func (c *Config) Exists(key Key) bool {
	key = sanitizeKey(key)

	_, ok := c.entries[key]

	return ok
}

// Get returns the configuration Value for the given Key.
// Returns an error if no values exists for the given Key.
func (c *Config) Get(k Key) (Value, error) {
	k = sanitizeKey(k)

	value, ok := c.entries[k]

	if !ok {
		return "", fmt.Errorf("value for %s does not exist", k)
	}

	return value, nil
}

// GetAll returns a map of all Key-Value-pairs
func (c *Config) GetAll() Entries {
	return maps.Clone(c.entries)
}

// GetChangeHistory returns a slice of all changes made to the config
func (c *Config) GetChangeHistory() []Change {
	return slices.Clone(c.changeHistory)
}

// Delete removes the configuration Key and Value
func (c *Config) Delete(k Key) {
	k = sanitizeKey(k)

	for configKey := range c.entries {
		if configKey == k {
			delete(c.entries, k)
			c.changeHistory = append(c.changeHistory, Change{KeyPath: k, Deleted: true})
		}
	}
}

// DeleteRecursive removes all configuration for the given Key, including all configuration for sub-keys
func (c *Config) DeleteRecursive(k Key) {
	k = sanitizeKey(k)

	c.Delete(k)

	//scan for subkeys
	if k != "" && !strings.HasSuffix(k.String(), keySeparator) {
		k = k + keySeparator
	}

	for configKey := range c.entries {
		if strings.HasPrefix(configKey.String(), k.String()) {
			delete(c.entries, configKey)
			c.changeHistory = append(c.changeHistory, Change{KeyPath: configKey, Deleted: true})
		}
	}
}

func (c *Config) DeleteAll() {
	// delete recursive from root
	c.DeleteRecursive(keySeparator)
}

func sanitizeKey(key Key) Key {
	sKey := key.String()

	if strings.HasPrefix(sKey, keySeparator) {
		return Key(sKey[1:])
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
