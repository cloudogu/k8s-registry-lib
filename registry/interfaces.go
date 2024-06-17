package registry

import "context"

// ConfigurationRegistry is able to manage the configuration of a single context
type ConfigurationRegistry interface {
	ConfigurationReader
	ConfigurationWriter
	ConfigurationWatcher
}

// ConfigurationReader is able to read the configuration of a single context
type ConfigurationReader interface {
	// Get returns a configuration value from the current context
	Get(ctx context.Context, key string) (string, error)
	// GetAll returns a map of key value pairs
	GetAll(ctx context.Context) (map[string]string, error)
	// Exists returns true if configuration key exists in the current context
	Exists(ctx context.Context, key string) (bool, error)
}

// ConfigurationWriter is able to write and delete the configuration of a single context
type ConfigurationWriter interface {
	// Set sets a configuration value in current context
	Set(ctx context.Context, key, value string) error
	// Delete removes a configuration key and value from the current context
	Delete(ctx context.Context, key string) error
	// DeleteRecursive removes a configuration key or directory from the current context
	DeleteRecursive(ctx context.Context, key string) error
	// DeleteAll removes all configuration keys
	DeleteAll(ctx context.Context) error
}

// ConfigurationWatcher is able to watch the configuration-changes
type ConfigurationWatcher interface {
	// Watch watches for changes of the provided config-key and sends the event through the channel
	Watch(ctx context.Context, key string, recursive bool) (ConfigWatch, error)
}
