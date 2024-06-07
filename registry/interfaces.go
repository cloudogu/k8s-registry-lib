package registry

import "context"

// ConfigurationRegistry is able to manage the configuration of a single context
type ConfigurationRegistry interface {
	// Set sets a configuration value in current context
	Set(ctx context.Context, key, value string) error
	// Get returns a configuration value from the current context
	Get(ctx context.Context, key string) (string, error)
	// GetAll returns a map of key value pairs
	GetAll(ctx context.Context) (map[string]string, error)
	// Delete removes a configuration key and value from the current context
	Delete(ctx context.Context, key string) error
	// DeleteRecursive removes a configuration key or directory from the current context
	DeleteRecursive(ctx context.Context, key string) error
	// Exists returns true if configuration key exists in the current context
	Exists(ctx context.Context, key string) (bool, error)
	// RemoveAll remove all configuration keys
	RemoveAll(ctx context.Context) error
	// GetOrFalse return false and empty string when the configuration value does not exist.
	// Otherwise, return true and the configuration value, even when the configuration value is an empty string.
	GetOrFalse(ctx context.Context, key string) (bool, string, error)
}
