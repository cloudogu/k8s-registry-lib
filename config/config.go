package config

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// SimpleDoguName represents a simple Dogu name as a string.
type SimpleDoguName string

// String returns the string representation of the SimpleDoguName.
func (s SimpleDoguName) String() string {
	return string(s)
}

// Key represents a configuration key as a string.
type Key string

// String returns the string representation of the Key.
func (k Key) String() string {
	return string(k)
}

// Value represents a configuration value as a string.
type Value string

// String returns the string representation of the Value.
func (v Value) String() string {
	return string(v)
}

// OptionalValue represents a configuration value as a string that may be null.
type OptionalValue struct {
	String string
	Exists bool // Exists is true if String is not NULL
}

// Entries represents a map of configuration keys and values.
type Entries map[Key]Value

// Change represents a change to a configuration key. When the Key has been deleted, Deleted is set to true.
type Change struct {
	KeyPath Key
	Deleted bool
}

// DiffResult represents a result of a configuration comparison.
type DiffResult struct {
	Key        Key
	Value      OptionalValue
	OtherValue OptionalValue
}

// Config represents a general configuration with entries and change history.
// PersistenceContext is used by a repository to detect conflicts due to remote changes.
type Config struct {
	entries             Entries
	changeHistory       []Change
	PersistenceContext  any
	ListResourceVersion string
}

type ConfigOption func(config *Config)

func WithPersistenceContext(pCtx any) ConfigOption {
	return func(config *Config) {
		config.PersistenceContext = pCtx
	}
}

// WithListResourceVersion adds the resourceVersion of the list containing the config. It's main use case is for the
// Config-Watches that operate on lists instead of single objects.
func WithListResourceVersion(resourceVersion string) ConfigOption {
	return func(config *Config) {
		config.ListResourceVersion = resourceVersion
	}
}

// CreateConfig creates a new configuration with the provided entries.
func CreateConfig(data Entries, options ...ConfigOption) Config {
	cfg := Config{
		entries:       data,
		changeHistory: make([]Change, 0),
	}

	for _, o := range options {
		o(&cfg)
	}

	return cfg
}

// Set sets the value for the given key in the configuration.
// Returns a new Config with the value set.
// Returns an error if the key is invalid or conflicts with existing keys.
func (c Config) Set(k Key, v Value) (Config, error) {
	k = sanitizeKey(k)

	if k == "" || k == keySeparator {
		return Config{}, fmt.Errorf("key is empty")
	}

	if strings.HasSuffix(k.String(), keySeparator) {
		return Config{}, fmt.Errorf("key %s must not be a dictionary", k)
	}

	subKey := k + keySeparator

	for configKey := range c.entries {
		if strings.HasPrefix(configKey.String(), subKey.String()) {
			return Config{}, fmt.Errorf("key %s is already used as dictionary", configKey)
		}
	}

	if lErr := validateNoDictionaryHasValue(k, k, c.entries); lErr != nil {
		return Config{}, fmt.Errorf("dictionary with key %s already has a value: %w", k, lErr)
	}

	c.entries[k] = v
	c.changeHistory = append(c.changeHistory, Change{KeyPath: k, Deleted: false})

	return c.createCopy(), nil
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

// Get returns the configuration value for the given key.
// When the key does not exist false is returned.
func (c Config) Get(k Key) (Value, bool) {
	k = sanitizeKey(k)

	v, ok := c.entries[k]

	return v, ok
}

// GetListResourceVersion returns the resourceVersion of the list containing the config. It's main use case is for the
// Config-Watches that operate on lists instead of single objects.
//func (c Config) GetListResourceVersion() string {
//	return c.listResourceVersion
//}

// SetListResourceVersion sets the resourceVersion of the list containing the config. It's main use case is for the
// Config-Watches that operate on lists instead of single objects.
//func (c Config) SetListResourceVersion(newResourceVersion string) {
//	c.listResourceVersion = newResourceVersion
//}

// GetAll returns a map of all Key-Value-pairs
func (c Config) GetAll() Entries {
	return maps.Clone(c.entries)
}

// GetChangeHistory returns a slice of all changes made to the configuration.
func (c Config) GetChangeHistory() []Change {
	return slices.Clone(c.changeHistory)
}

// Delete removes the configuration Key and Value.
// Returns a new Config with the Key deleted.
func (c Config) Delete(k Key) Config {
	k = sanitizeKey(k)

	for configKey := range c.entries {
		if configKey == k {
			delete(c.entries, k)
			c.changeHistory = append(c.changeHistory, Change{KeyPath: k, Deleted: true})
		}
	}

	return c.createCopy()
}

// DeleteRecursive removes all configuration for the given Key, including all configuration for sub-keys.
// Returns a new Config with without given Key.
func (c Config) DeleteRecursive(k Key) Config {
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

	return c.createCopy()
}

// DeleteAll removes all key values pairs for the configuration.
// Returns a new empty Config with a change history containing all keys that haven been deleted.
func (c Config) DeleteAll() Config {
	// delete recursive from root
	c.DeleteRecursive(keySeparator)

	return Config{
		entries:            make(Entries),
		changeHistory:      slices.Clone(c.changeHistory),
		PersistenceContext: c.PersistenceContext,
	}
}

// Diff returns a list of DiffResult with all values that differs for a given key.
func (c Config) Diff(other Config) []DiffResult {
	m := make(map[Key]DiffResult, len(c.entries))

	for k, v := range c.entries {
		m[k] = DiffResult{
			Key:        k,
			Value:      OptionalValue{String: v.String(), Exists: true},
			OtherValue: OptionalValue{Exists: false},
		}
	}

	for kOther, vOther := range other.entries {
		mod, ok := m[kOther]
		if !ok {
			m[kOther] = DiffResult{
				Key:        kOther,
				Value:      OptionalValue{Exists: false},
				OtherValue: OptionalValue{String: vOther.String(), Exists: true},
			}

			continue
		}

		m[kOther] = DiffResult{
			Key:        kOther,
			Value:      mod.Value,
			OtherValue: OptionalValue{String: vOther.String(), Exists: true},
		}
	}

	mods := make([]DiffResult, 0, len(m))
	for _, v := range m {
		if v.Value.Exists != v.OtherValue.Exists || v.Value.String != v.OtherValue.String {
			mods = append(mods, v)
		}
	}

	return mods
}

func (c Config) createCopy() Config {
	return Config{
		entries:            maps.Clone(c.entries),
		changeHistory:      slices.Clone(c.changeHistory),
		PersistenceContext: c.PersistenceContext,
	}
}

func sanitizeKey(key Key) Key {
	sKey := key.String()

	if strings.HasPrefix(sKey, keySeparator) {
		return Key(sKey[1:])
	}

	return key
}
